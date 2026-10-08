package engine

import (
	"testing"
	"time"

	"IDIG4110/rules-engine/internal/rule"
	"IDIG4110/shared/dto"
)

const gw = "11111111-1111-1111-1111-111111111111"

var t0 = time.Date(2026, 10, 6, 12, 0, 0, 0, time.UTC)

func num(entity string, v float64, at time.Time) dto.Reading {
	return dto.Reading{GatewayID: gw, ExternalEntityID: entity, Timestamp: at, ValueNum: &v}
}

func text(entity, v string, at time.Time) dto.Reading {
	return dto.Reading{GatewayID: gw, ExternalEntityID: entity, Timestamp: at, ValueText: &v}
}

func newRule(id string, trigger rule.Trigger, conditions ...rule.Match) rule.Rule {
	return rule.Rule{
		ID: id, Name: id, Enabled: true, UpdatedAt: t0,
		Trigger: trigger, Conditions: conditions,
		Actions: []rule.Action{{Type: rule.ActionIncident, Severity: "warning", Message: id}},
	}
}

func above(entity string, v float64) rule.Trigger {
	return rule.Trigger{Match: rule.Match{Entity: entity, Operator: rule.OpGt, Value: v}}
}

func engineWith(rules ...rule.Rule) *Engine {
	e := New()
	e.SetRules(rules)
	return e
}

func ids(fs []Firing) []string {
	out := make([]string, len(fs))
	for i, f := range fs {
		out[i] = f.Rule.ID
	}
	return out
}

func expectFired(t *testing.T, got []Firing, want ...string) {
	t.Helper()
	g := ids(got)
	if len(g) != len(want) {
		t.Fatalf("fired %v, want %v", g, want)
	}
	for i := range want {
		if g[i] != want[i] {
			t.Fatalf("fired %v, want %v", g, want)
		}
	}
}

func TestThresholdFiresOncePerRisingEdge(t *testing.T) {
	e := engineWith(newRule("hot", above("sensor.temp", 30)))

	expectFired(t, e.Handle(num("sensor.temp", 25, t0), t0))
	expectFired(t, e.Handle(num("sensor.temp", 31, t0.Add(time.Second)), t0.Add(time.Second)), "hot")
	// Still above: no repeat while the rule stays true.
	expectFired(t, e.Handle(num("sensor.temp", 35, t0.Add(2*time.Second)), t0.Add(2*time.Second)))
	// Back below, then above again: a new edge fires again.
	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(3*time.Second)), t0.Add(3*time.Second)))
	expectFired(t, e.Handle(num("sensor.temp", 32, t0.Add(4*time.Second)), t0.Add(4*time.Second)), "hot")
}

func TestTextStateAndBooleanValue(t *testing.T) {
	smoke := newRule("smoke", rule.Trigger{Match: rule.Match{Entity: "binary_sensor.smoke", Operator: rule.OpEq, Value: true}})
	door := newRule("door", rule.Trigger{Match: rule.Match{Entity: "lock.front", Operator: rule.OpEq, Value: "unlocked"}})
	e := engineWith(smoke, door)

	expectFired(t, e.Handle(text("binary_sensor.smoke", "off", t0), t0))
	expectFired(t, e.Handle(text("binary_sensor.smoke", "ON", t0), t0), "smoke")
	expectFired(t, e.Handle(text("lock.front", "unlocked", t0), t0), "door")
}

func TestNumberSentAsText(t *testing.T) {
	e := engineWith(newRule("hot", above("sensor.temp", 30)))
	expectFired(t, e.Handle(text("sensor.temp", "30.5", t0), t0), "hot")
}

func TestConditionsMustAllHold(t *testing.T) {
	occupied := rule.Match{Entity: "binary_sensor.occupied", Operator: rule.OpEq, Value: "on"}
	e := engineWith(newRule("hot-home", above("sensor.temp", 30), occupied))

	// No reading for the condition entity yet: does not hold.
	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0))
	// Condition becomes true while the trigger already holds: the rule fires.
	expectFired(t, e.Handle(text("binary_sensor.occupied", "on", t0), t0), "hot-home")
	expectFired(t, e.Handle(text("binary_sensor.occupied", "off", t0), t0))
	expectFired(t, e.Handle(text("binary_sensor.occupied", "on", t0), t0), "hot-home")
}

func TestForSecondsWaitsAndRequiresTheRuleToKeepHolding(t *testing.T) {
	trigger := above("sensor.temp", 30)
	trigger.ForSeconds = 60
	e := engineWith(newRule("sustained", trigger))

	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0))
	expectFired(t, e.Tick(t0.Add(59*time.Second)))
	expectFired(t, e.Tick(t0.Add(60*time.Second)), "sustained")
	expectFired(t, e.Tick(t0.Add(120*time.Second))) // fires once

	// Dropping below before the time is up cancels it.
	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(130*time.Second)), t0.Add(130*time.Second)))
	expectFired(t, e.Handle(num("sensor.temp", 31, t0.Add(140*time.Second)), t0.Add(140*time.Second)))
	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(150*time.Second)), t0.Add(150*time.Second)))
	expectFired(t, e.Tick(t0.Add(300*time.Second)))
}

func TestCooldownSuppressesRefiring(t *testing.T) {
	r := newRule("hot", above("sensor.temp", 30))
	r.CooldownSeconds = 300
	e := engineWith(r)

	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0), "hot")
	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(10*time.Second)), t0.Add(10*time.Second)))
	expectFired(t, e.Handle(num("sensor.temp", 31, t0.Add(20*time.Second)), t0.Add(20*time.Second)))
	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(30*time.Second)), t0.Add(30*time.Second)))
	at := t0.Add(301 * time.Second)
	expectFired(t, e.Handle(num("sensor.temp", 31, at), at), "hot")
}

func TestPriorityOrderAndDisabledRules(t *testing.T) {
	low := newRule("low", above("sensor.temp", 30))
	high := newRule("high", above("sensor.temp", 30))
	high.Priority = 10
	off := newRule("off", above("sensor.temp", 30))
	off.Enabled = false
	e := engineWith(low, off, high)

	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0), "high", "low")
}

func TestOlderReadingsAreIgnored(t *testing.T) {
	e := engineWith(newRule("hot", above("sensor.temp", 30)))

	expectFired(t, e.Handle(num("sensor.temp", 20, t0.Add(time.Minute)), t0))
	// A late reading from before the known one must not count.
	expectFired(t, e.Handle(num("sensor.temp", 40, t0), t0))
}

func TestGatewayScopedMatch(t *testing.T) {
	other := "22222222-2222-2222-2222-222222222222"
	trigger := above("sensor.temp", 30)
	trigger.GatewayID = other
	e := engineWith(newRule("other-gw", trigger))

	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0)) // gateway gw, not other
	r := num("sensor.temp", 31, t0)
	r.GatewayID = other
	expectFired(t, e.Handle(r, t0), "other-gw")
}

func TestSetRulesResetsChangedRulesButKeepsCooldown(t *testing.T) {
	r := newRule("hot", above("sensor.temp", 30))
	r.CooldownSeconds = 300
	e := engineWith(r)
	expectFired(t, e.Handle(num("sensor.temp", 31, t0), t0), "hot")

	// Unchanged rule: still active, so no new edge.
	e.SetRules([]rule.Rule{r})
	expectFired(t, e.Handle(num("sensor.temp", 32, t0.Add(time.Second)), t0.Add(time.Second)))

	// Edited rule starts inactive, but the cooldown still applies.
	r.UpdatedAt = t0.Add(time.Minute)
	e.SetRules([]rule.Rule{r})
	expectFired(t, e.Handle(num("sensor.temp", 33, t0.Add(2*time.Second)), t0.Add(2*time.Second)))
	at := t0.Add(400 * time.Second)
	e.SetRules([]rule.Rule{func() rule.Rule { r.UpdatedAt = at; return r }()})
	expectFired(t, e.Handle(num("sensor.temp", 34, at), at), "hot")
}
