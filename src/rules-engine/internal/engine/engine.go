// Package engine evaluates rules against the stream of readings.
//
// It keeps the latest reading of every entity, so conditions can look at
// other sensors. Rules are edge-triggered: a rule fires when it goes from
// false to true, not on every reading while it stays true. A trigger with
// for_seconds fires only after the rule has held that long (checked by Tick),
// and a cooldown suppresses firing again too soon.
//
// The engine does no I/O and takes the current time as an argument, so it can
// be tested deterministically. It is safe for concurrent use.
package engine

import (
	"sort"
	"sync"
	"time"

	"IDIG4110/rules-engine/internal/rule"
	"IDIG4110/shared/dto"
)

// Firing is one rule firing, with the trigger entity's reading that made the
// rule true.
type Firing struct {
	Rule    rule.Rule
	Reading dto.Reading
	At      time.Time
}

type ruleState struct {
	updatedAt time.Time // rule version this state belongs to
	active    bool      // the rule evaluated true last time
	pending   bool      // true, waiting for for_seconds to pass
	since     time.Time // when the rule became true
	lastFired time.Time
}

type Engine struct {
	mu       sync.Mutex
	rules    []rule.Rule // enabled rules, highest priority first
	states   map[string]*ruleState
	byKey    map[string]dto.Reading // latest reading per gateway/entity
	byEntity map[string]dto.Reading // latest reading per entity, any gateway
}

func New() *Engine {
	return &Engine{
		states:   map[string]*ruleState{},
		byKey:    map[string]dto.Reading{},
		byEntity: map[string]dto.Reading{},
	}
}

// SetRules replaces the rule set. Disabled rules are dropped. A rule whose
// definition changed starts over as inactive; its cooldown is kept.
func (e *Engine) SetRules(rules []rule.Rule) {
	e.mu.Lock()
	defer e.mu.Unlock()

	enabled := make([]rule.Rule, 0, len(rules))
	states := make(map[string]*ruleState, len(rules))
	for _, r := range rules {
		if !r.Enabled {
			continue
		}
		enabled = append(enabled, r)
		st, ok := e.states[r.ID]
		switch {
		case !ok:
			st = &ruleState{updatedAt: r.UpdatedAt}
		case !st.updatedAt.Equal(r.UpdatedAt):
			st = &ruleState{updatedAt: r.UpdatedAt, lastFired: st.lastFired}
		}
		states[r.ID] = st
	}
	sort.SliceStable(enabled, func(i, j int) bool { return enabled[i].Priority > enabled[j].Priority })
	e.rules, e.states = enabled, states
}

// Handle records reading r and returns the rules that fire because of it, in
// priority order. A reading older than the one already known for its entity
// is ignored, so late or replayed data never overwrites newer state.
func (e *Engine) Handle(r dto.Reading, now time.Time) []Firing {
	e.mu.Lock()
	defer e.mu.Unlock()

	if known, ok := e.byKey[r.Key()]; ok && r.Timestamp.Before(known.Timestamp) {
		return nil
	}
	e.byKey[r.Key()] = r
	if known, ok := e.byEntity[r.ExternalEntityID]; !ok || !r.Timestamp.Before(known.Timestamp) {
		e.byEntity[r.ExternalEntityID] = r
	}

	var fired []Firing
	for _, ru := range e.rules {
		if !refersTo(ru, r) {
			continue
		}
		if f := e.step(ru, now); f != nil {
			fired = append(fired, *f)
		}
	}
	return fired
}

// Tick fires rules whose for_seconds has passed and that still hold.
func (e *Engine) Tick(now time.Time) []Firing {
	e.mu.Lock()
	defer e.mu.Unlock()

	var fired []Firing
	for _, ru := range e.rules {
		st := e.states[ru.ID]
		if !st.pending || now.Sub(st.since) < seconds(ru.Trigger.ForSeconds) {
			continue
		}
		st.pending = false
		reading, holds := e.evaluate(ru)
		if !holds {
			st.active = false
			continue
		}
		if f := e.fire(ru, st, reading, now); f != nil {
			fired = append(fired, *f)
		}
	}
	return fired
}

// step re-evaluates one rule after a relevant reading.
func (e *Engine) step(ru rule.Rule, now time.Time) *Firing {
	st := e.states[ru.ID]
	reading, holds := e.evaluate(ru)
	if !holds {
		st.active, st.pending = false, false
		return nil
	}
	if st.active {
		return nil // already fired or waiting for for_seconds
	}
	st.active, st.since = true, now
	if ru.Trigger.ForSeconds > 0 {
		st.pending = true
		return nil
	}
	return e.fire(ru, st, reading, now)
}

func (e *Engine) fire(ru rule.Rule, st *ruleState, reading dto.Reading, now time.Time) *Firing {
	if !st.lastFired.IsZero() && now.Sub(st.lastFired) < seconds(ru.CooldownSeconds) {
		return nil
	}
	st.lastFired = now
	return &Firing{Rule: ru, Reading: reading, At: now}
}

// evaluate checks the trigger and all conditions against the latest known
// readings. An entity with no reading yet counts as not matching.
func (e *Engine) evaluate(ru rule.Rule) (dto.Reading, bool) {
	trigger, ok := e.latest(ru.Trigger.Match)
	if !ok || !ru.Trigger.Matches(trigger) {
		return dto.Reading{}, false
	}
	for _, c := range ru.Conditions {
		r, ok := e.latest(c)
		if !ok || !c.Matches(r) {
			return dto.Reading{}, false
		}
	}
	return trigger, true
}

func (e *Engine) latest(m rule.Match) (dto.Reading, bool) {
	if m.GatewayID != "" {
		r, ok := e.byKey[dto.EntityKey(m.GatewayID, m.Entity)]
		return r, ok
	}
	r, ok := e.byEntity[m.Entity]
	return r, ok
}

func refersTo(ru rule.Rule, r dto.Reading) bool {
	if ru.Trigger.Refers(r) {
		return true
	}
	for _, c := range ru.Conditions {
		if c.Refers(r) {
			return true
		}
	}
	return false
}

func seconds(n int) time.Duration {
	return time.Duration(n) * time.Second
}
