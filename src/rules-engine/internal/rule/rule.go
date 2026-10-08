// Package rule defines automation rules: a trigger, optional conditions and
// a list of actions. See docs/architecture/rules-engine.md.
package rule

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"time"
)

type Operator string

const (
	OpEq  Operator = "eq"
	OpNe  Operator = "ne"
	OpGt  Operator = "gt"
	OpGte Operator = "gte"
	OpLt  Operator = "lt"
	OpLte Operator = "lte"
)

// Match compares the latest reading of one entity against a value. An empty
// GatewayID matches the entity on any gateway.
type Match struct {
	GatewayID string   `json:"gateway_id,omitempty"`
	Entity    string   `json:"entity"`
	Operator  Operator `json:"operator"`
	Value     any      `json:"value"`
}

// Trigger is the match that makes a rule fire. With ForSeconds > 0 the rule
// only fires once the trigger and all conditions have held that long.
type Trigger struct {
	Match
	ForSeconds int `json:"for_seconds,omitempty"`
}

type ActionType string

const (
	ActionCommand  ActionType = "command"
	ActionIncident ActionType = "incident"
	ActionNotify   ActionType = "notify"
)

// Action is one thing to do when a rule fires. Which fields apply depends on
// Type:
//   - command:  Entity, Command, optional GatewayID (defaults to the gateway
//     of the triggering reading) and Parameters
//   - incident: Severity and Message ({{entity}} and {{value}} are replaced)
//   - notify:   Channels; delivered for the rule's incident
type Action struct {
	Type       ActionType     `json:"type"`
	GatewayID  string         `json:"gateway_id,omitempty"`
	Entity     string         `json:"entity,omitempty"`
	Command    string         `json:"command,omitempty"`
	Parameters map[string]any `json:"parameters,omitempty"`
	Severity   string         `json:"severity,omitempty"`
	Message    string         `json:"message,omitempty"`
	Channels   []string       `json:"channels,omitempty"`
}

type Rule struct {
	ID              string     `json:"id"`
	HomeID          string     `json:"home_id"`
	Name            string     `json:"name"`
	Description     string     `json:"description"`
	Enabled         bool       `json:"enabled"`
	Priority        int        `json:"priority"`
	Trigger         Trigger    `json:"trigger"`
	Conditions      []Match    `json:"conditions"`
	CooldownSeconds int        `json:"cooldown_seconds"`
	Actions         []Action   `json:"actions"`
	FireCount       int        `json:"fire_count"`
	LastFiredAt     *time.Time `json:"last_fired_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Channels returns the notification channels of all notify actions, never nil.
func (r Rule) Channels() []string {
	channels := []string{}
	for _, a := range r.Actions {
		if a.Type == ActionNotify {
			channels = append(channels, a.Channels...)
		}
	}
	return channels
}

// --- storage --------------------------------------------------------------

// triggerConfig and actionConfig are the shapes stored in the rules table's
// trigger_config and action_config columns.
type triggerConfig struct {
	Trigger         Trigger `json:"trigger"`
	Conditions      []Match `json:"conditions"`
	CooldownSeconds int     `json:"cooldown_seconds"`
}

type actionConfig struct {
	Actions []Action `json:"actions"`
}

func (r Rule) TriggerConfig() ([]byte, error) {
	return json.Marshal(triggerConfig{Trigger: r.Trigger, Conditions: r.Conditions, CooldownSeconds: r.CooldownSeconds})
}

func (r Rule) ActionConfig() ([]byte, error) {
	return json.Marshal(actionConfig{Actions: r.Actions})
}

// ApplyConfigs fills the trigger, conditions, cooldown and actions of r from
// the stored JSON columns.
func (r *Rule) ApplyConfigs(triggerJSON, actionJSON []byte) error {
	var tc triggerConfig
	if err := json.Unmarshal(triggerJSON, &tc); err != nil {
		return fmt.Errorf("rule %s: trigger_config: %w", r.ID, err)
	}
	var ac actionConfig
	if err := json.Unmarshal(actionJSON, &ac); err != nil {
		return fmt.Errorf("rule %s: action_config: %w", r.ID, err)
	}
	r.Trigger, r.Conditions, r.CooldownSeconds, r.Actions = tc.Trigger, tc.Conditions, tc.CooldownSeconds, ac.Actions
	return nil
}

// --- validation -----------------------------------------------------------

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

var (
	severities = map[string]bool{"info": true, "warning": true, "critical": true}
	channels   = map[string]bool{"ui": true, "email": true, "push": true}
)

// Validate reports every problem with r, joined into one error.
func (r Rule) Validate() error {
	var errs []error
	if !uuidPattern.MatchString(r.HomeID) {
		errs = append(errs, errors.New("home_id must be a UUID"))
	}
	if r.Name == "" || len(r.Name) > 255 {
		errs = append(errs, errors.New("name is required and at most 255 characters"))
	}
	if err := r.Trigger.Match.validate("trigger"); err != nil {
		errs = append(errs, err)
	}
	if r.Trigger.ForSeconds < 0 {
		errs = append(errs, errors.New("trigger.for_seconds must not be negative"))
	}
	if r.CooldownSeconds < 0 {
		errs = append(errs, errors.New("cooldown_seconds must not be negative"))
	}
	for i, c := range r.Conditions {
		if err := c.validate(fmt.Sprintf("conditions[%d]", i)); err != nil {
			errs = append(errs, err)
		}
	}

	if len(r.Actions) == 0 {
		errs = append(errs, errors.New("at least one action is required"))
	}
	incidents, notifies := 0, 0
	for i, a := range r.Actions {
		where := fmt.Sprintf("actions[%d]", i)
		switch a.Type {
		case ActionCommand:
			if a.Entity == "" || a.Command == "" {
				errs = append(errs, fmt.Errorf("%s: command needs entity and command", where))
			}
			if a.GatewayID != "" && !uuidPattern.MatchString(a.GatewayID) {
				errs = append(errs, fmt.Errorf("%s: gateway_id must be a UUID", where))
			}
		case ActionIncident:
			incidents++
			if !severities[a.Severity] {
				errs = append(errs, fmt.Errorf("%s: severity must be info, warning or critical", where))
			}
			if a.Message == "" {
				errs = append(errs, fmt.Errorf("%s: incident needs a message", where))
			}
		case ActionNotify:
			notifies++
			if len(a.Channels) == 0 {
				errs = append(errs, fmt.Errorf("%s: notify needs at least one channel", where))
			}
			for _, ch := range a.Channels {
				if !channels[ch] {
					errs = append(errs, fmt.Errorf("%s: unknown channel %q (ui, email, push)", where, ch))
				}
			}
		default:
			errs = append(errs, fmt.Errorf("%s: unknown action type %q (command, incident, notify)", where, a.Type))
		}
	}
	if incidents > 1 {
		errs = append(errs, errors.New("at most one incident action per rule"))
	}
	// Notifications are delivered per incident (notifications.incident_id).
	if notifies > 0 && incidents == 0 {
		errs = append(errs, errors.New("notify needs an incident action in the same rule"))
	}
	return errors.Join(errs...)
}

func (m Match) validate(where string) error {
	if m.Entity == "" {
		return fmt.Errorf("%s: entity is required", where)
	}
	if m.GatewayID != "" && !uuidPattern.MatchString(m.GatewayID) {
		return fmt.Errorf("%s: gateway_id must be a UUID", where)
	}
	switch m.Operator {
	case OpEq, OpNe:
		switch m.Value.(type) {
		case string, float64, bool:
		default:
			return fmt.Errorf("%s: value must be a string, number or boolean", where)
		}
	case OpGt, OpGte, OpLt, OpLte:
		if _, ok := m.Value.(float64); !ok {
			return fmt.Errorf("%s: operator %s needs a numeric value", where, m.Operator)
		}
	default:
		return fmt.Errorf("%s: unknown operator %q (eq, ne, gt, gte, lt, lte)", where, m.Operator)
	}
	return nil
}
