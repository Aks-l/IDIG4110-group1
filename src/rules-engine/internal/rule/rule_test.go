package rule

import (
	"strings"
	"testing"
)

func validRule() Rule {
	return Rule{
		HomeID: "00000000-0000-0000-0000-000000000001",
		Name:   "Smoke in kitchen",
		Trigger: Trigger{Match: Match{
			Entity: "binary_sensor.kitchen_smoke", Operator: OpEq, Value: true,
		}},
		Actions: []Action{
			{Type: ActionCommand, Entity: "lock.front_door", Command: "unlock"},
			{Type: ActionIncident, Severity: "critical", Message: "Smoke detected"},
			{Type: ActionNotify, Channels: []string{"ui", "push"}},
		},
	}
}

func TestValidRulePasses(t *testing.T) {
	if err := validRule().Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateReportsProblems(t *testing.T) {
	cases := map[string]struct {
		edit func(*Rule)
		want string
	}{
		"bad home":            {func(r *Rule) { r.HomeID = "home-1" }, "home_id must be a UUID"},
		"no name":             {func(r *Rule) { r.Name = "" }, "name is required"},
		"unknown operator":    {func(r *Rule) { r.Trigger.Operator = "between" }, "unknown operator"},
		"gt needs number":     {func(r *Rule) { r.Trigger.Operator, r.Trigger.Value = OpGt, "hot" }, "needs a numeric value"},
		"negative for":        {func(r *Rule) { r.Trigger.ForSeconds = -1 }, "for_seconds"},
		"no actions":          {func(r *Rule) { r.Actions = nil }, "at least one action"},
		"bad severity":        {func(r *Rule) { r.Actions[1].Severity = "urgent" }, "severity"},
		"bad channel":         {func(r *Rule) { r.Actions[2].Channels = []string{"sms"} }, "unknown channel"},
		"notify w/o incident": {func(r *Rule) { r.Actions = append(r.Actions[:1], r.Actions[2]) }, "notify needs an incident"},
		"command w/o entity":  {func(r *Rule) { r.Actions[0].Entity = "" }, "command needs entity"},
		"condition entity":    {func(r *Rule) { r.Conditions = []Match{{Operator: OpEq, Value: "on"}} }, "conditions[0]: entity is required"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := validRule()
			tc.edit(&r)
			err := r.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("got %v, want error containing %q", err, tc.want)
			}
		})
	}
}

func TestConfigsRoundTrip(t *testing.T) {
	r := validRule()
	r.Trigger.ForSeconds = 30
	r.CooldownSeconds = 300
	tc, err := r.TriggerConfig()
	if err != nil {
		t.Fatal(err)
	}
	ac, err := r.ActionConfig()
	if err != nil {
		t.Fatal(err)
	}

	var got Rule
	if err := got.ApplyConfigs(tc, ac); err != nil {
		t.Fatal(err)
	}
	if got.Trigger.Entity != r.Trigger.Entity || got.Trigger.ForSeconds != 30 || got.CooldownSeconds != 300 || len(got.Actions) != 3 {
		t.Fatalf("round trip lost data: %+v", got)
	}
	if got.Trigger.Value != true {
		t.Fatalf("value = %#v, want true", got.Trigger.Value)
	}
}
