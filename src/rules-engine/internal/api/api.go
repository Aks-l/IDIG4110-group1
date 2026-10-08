// Package api is the REST API for rules and incidents, served under /api/v1
// and meant to be reached through the API gateway.
package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"

	"IDIG4110/rules-engine/internal/rule"
	"IDIG4110/rules-engine/internal/store"
	"IDIG4110/shared/httperror"
	jsonutils "IDIG4110/shared/json-utils"
)

const prefix = "/api/v1"

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type API struct {
	store *store.Store
	// reload pushes the current rules into the running engine after a change.
	reload func(context.Context) error
}

func New(s *store.Store, reload func(context.Context) error) *API {
	return &API{store: s, reload: reload}
}

func (a *API) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	mux.HandleFunc("GET "+prefix+"/rules", a.listRules)
	mux.HandleFunc("POST "+prefix+"/rules", a.createRule)
	mux.HandleFunc("GET "+prefix+"/rules/{id}", a.getRule)
	mux.HandleFunc("PUT "+prefix+"/rules/{id}", a.updateRule)
	mux.HandleFunc("PATCH "+prefix+"/rules/{id}", a.setRuleEnabled)
	mux.HandleFunc("DELETE "+prefix+"/rules/{id}", a.deleteRule)

	mux.HandleFunc("GET "+prefix+"/incidents", a.listIncidents)
	mux.HandleFunc("POST "+prefix+"/incidents/{id}/acknowledge", a.acknowledgeIncident)
	mux.HandleFunc("POST "+prefix+"/incidents/{id}/resolve", a.resolveIncident)
	return mux
}

// ruleInput is the body of POST and PUT /rules.
type ruleInput struct {
	HomeID          string        `json:"home_id"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Enabled         *bool         `json:"enabled"`
	Priority        int           `json:"priority"`
	Trigger         rule.Trigger  `json:"trigger"`
	Conditions      []rule.Match  `json:"conditions"`
	CooldownSeconds int           `json:"cooldown_seconds"`
	Actions         []rule.Action `json:"actions"`
}

func (in ruleInput) toRule(id string) rule.Rule {
	enabled := true
	if in.Enabled != nil {
		enabled = *in.Enabled
	}
	conditions := in.Conditions
	if conditions == nil {
		conditions = []rule.Match{}
	}
	return rule.Rule{
		ID: id, HomeID: in.HomeID, Name: in.Name, Description: in.Description,
		Enabled: enabled, Priority: in.Priority, Trigger: in.Trigger, Conditions: conditions,
		CooldownSeconds: in.CooldownSeconds, Actions: in.Actions,
	}
}

func (a *API) listRules(w http.ResponseWriter, r *http.Request) {
	rules, err := a.store.ListRules(r.Context(), r.URL.Query().Get("home_id"))
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, rules)
}

func (a *API) getRule(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	ru, err := a.store.GetRule(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, ru)
}

func (a *API) createRule(w http.ResponseWriter, r *http.Request) {
	in, err := jsonutils.Decode[ruleInput](r)
	if err != nil {
		httperror.HandleError(w, http.StatusBadRequest, err, "invalid JSON body")
		return
	}
	ru := in.toRule("")
	if err := ru.Validate(); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, err.Error())
		return
	}
	saved, err := a.store.CreateRule(r.Context(), ru)
	if err != nil {
		fail(w, err)
		return
	}
	a.reloadEngine(r.Context())
	respond(w, http.StatusCreated, saved)
}

func (a *API) updateRule(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	in, err := jsonutils.Decode[ruleInput](r)
	if err != nil {
		httperror.HandleError(w, http.StatusBadRequest, err, "invalid JSON body")
		return
	}
	ru := in.toRule(id)
	if err := ru.Validate(); err != nil {
		httperror.HandleError(w, http.StatusUnprocessableEntity, err, err.Error())
		return
	}
	saved, err := a.store.UpdateRule(r.Context(), ru)
	if err != nil {
		fail(w, err)
		return
	}
	a.reloadEngine(r.Context())
	respond(w, http.StatusOK, saved)
}

// setRuleEnabled handles PATCH /rules/{id} with {"enabled": true|false}, the
// call the automations page makes when a rule is toggled.
func (a *API) setRuleEnabled(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	body, err := jsonutils.Decode[struct {
		Enabled *bool `json:"enabled"`
	}](r)
	if err != nil || body.Enabled == nil {
		httperror.HandleError(w, http.StatusBadRequest, errors.New("missing enabled"), `body must be {"enabled": true|false}`)
		return
	}
	saved, err := a.store.SetRuleEnabled(r.Context(), id, *body.Enabled)
	if err != nil {
		fail(w, err)
		return
	}
	a.reloadEngine(r.Context())
	respond(w, http.StatusOK, saved)
}

func (a *API) deleteRule(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	if err := a.store.DeleteRule(r.Context(), id); err != nil {
		fail(w, err)
		return
	}
	a.reloadEngine(r.Context())
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) listIncidents(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit := 100
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > 1000 {
			httperror.HandleError(w, http.StatusBadRequest, errors.New("bad limit"), "limit must be 1-1000")
			return
		}
		limit = n
	}
	incidents, err := a.store.ListIncidents(r.Context(), q.Get("home_id"), q.Get("status"), limit)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, incidents)
}

func (a *API) acknowledgeIncident(w http.ResponseWriter, r *http.Request) {
	a.incidentTransition(w, r, a.store.AcknowledgeIncident)
}

func (a *API) resolveIncident(w http.ResponseWriter, r *http.Request) {
	a.incidentTransition(w, r, a.store.ResolveIncident)
}

func (a *API) incidentTransition(w http.ResponseWriter, r *http.Request, do func(context.Context, string) (store.Incident, error)) {
	id, ok := pathID(w, r)
	if !ok {
		return
	}
	inc, err := do(r.Context(), id)
	if err != nil {
		fail(w, err)
		return
	}
	respond(w, http.StatusOK, inc)
}

// reloadEngine applies a rule change at once. If it fails, the change is
// still saved and the periodic refresh picks it up.
func (a *API) reloadEngine(ctx context.Context) {
	if err := a.reload(ctx); err != nil {
		slog.Error("Reloading rules after change", "error", err)
	}
}

func pathID(w http.ResponseWriter, r *http.Request) (string, bool) {
	id := r.PathValue("id")
	if !uuidPattern.MatchString(id) {
		httperror.HandleError(w, http.StatusNotFound, errors.New("bad id"), httperror.ErrNotFound)
		return "", false
	}
	return id, true
}

func fail(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		httperror.HandleError(w, http.StatusNotFound, err, httperror.ErrNotFound)
	case errors.Is(err, store.ErrConflict):
		httperror.HandleError(w, http.StatusConflict, err, err.Error())
	default:
		httperror.HandleError(w, http.StatusInternalServerError, err, httperror.ErrInternalServerError)
	}
}

func respond(w http.ResponseWriter, code int, v any) {
	if err := jsonutils.Encode(w, code, v); err != nil {
		slog.Error("Writing response", "error", err)
	}
}
