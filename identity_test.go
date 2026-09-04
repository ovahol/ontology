package ontology

import "testing"

func testOvaholConventions() Conventions {
	return Conventions{
		UnknownDevice:       "Unknown Device",
		UnknownBrand:        "Unknown Brand",
		UnknownManufacturer: "Unknown Manufacturer",
		UnknownModelPrefix:  "Unknown Model - ",
		Statuses: []string{
			"In-Service",
			"Decommissioned",
			"Transferred",
			"Standby / Spare",
			"Under Maintenance",
			"Out of Service",
			"Disposed",
			"New / Commissioning",
		},
		StatusSynonyms: map[string][]string{
			"In-Service":        {"functional", "functioning", "active", "in active service", "working"},
			"Under Maintenance": {"faulty", "not working", "broken down", "broken down and repairable"},
		},
		DefaultStatus: "New / Commissioning",
	}
}

func TestReconcileIdentityAppliesUnknownConventions(t *testing.T) {
	conv := testOvaholConventions()
	out := ReconcileIdentity(IdentityInput{}, nil, conv)

	if out.Device != "Unknown Device" {
		t.Errorf("Device = %q, want Unknown Device", out.Device)
	}
	if out.Model != "Unknown Model - Unknown Device" {
		t.Errorf("Model = %q, want 'Unknown Model - Unknown Device'", out.Model)
	}
	if out.Brand != "Unknown Brand" {
		t.Errorf("Brand = %q, want Unknown Brand", out.Brand)
	}
	if out.Manufacturer != "Unknown Manufacturer" {
		t.Errorf("Manufacturer = %q, want Unknown Manufacturer", out.Manufacturer)
	}
	if out.Status != "New / Commissioning" {
		t.Errorf("empty status = %q, want default New / Commissioning", out.Status)
	}
}

func TestReconcileIdentityUnknownModelEmbedsDevice(t *testing.T) {
	conv := testOvaholConventions()
	out := ReconcileIdentity(IdentityInput{DeviceName: "Infusion pump"}, nil, conv)
	if want := "Unknown Model - Infusion pump"; out.Model != want {
		t.Errorf("Model = %q, want %q", out.Model, want)
	}
}

func TestReconcileIdentityKeepsResolvedFields(t *testing.T) {
	conv := testOvaholConventions()
	out := ReconcileIdentity(IdentityInput{
		DeviceName:   "Infusion pump",
		Model:        "Volumetric 10",
		Brand:        "Bbraun",
		Manufacturer: "B. Braun Melsungen AG",
		Status:       "In-Service",
	}, nil, conv)
	if out.Model != "Volumetric 10" || out.Brand != "Bbraun" || out.Manufacturer != "B. Braun Melsungen AG" {
		t.Errorf("resolved fields clobbered: %+v", out)
	}
	if out.Status != "In-Service" {
		t.Errorf("member status = %q, want In-Service", out.Status)
	}
}

type testResolver struct {
	IdentityResult
}

func (r testResolver) Resolve(in IdentityInput) (IdentityResult, bool) {
	return r.IdentityResult, true
}

func TestReconcileIdentityUsesResolver(t *testing.T) {
	conv := testOvaholConventions()
	res := testResolver{IdentityResult{
		Model:        "Canonical Model",
		Brand:        "Canonical Brand",
		Manufacturer: "Canonical Manufacturer",
	}}
	out := ReconcileIdentity(IdentityInput{Brand: "garbled"}, res, conv)
	if out.Model != "Canonical Model" || out.Brand != "Canonical Brand" || out.Manufacturer != "Canonical Manufacturer" {
		t.Errorf("resolver values not applied: %+v", out)
	}
}

func TestReconcileIdentityStatusNormalization(t *testing.T) {
	conv := testOvaholConventions()
	cases := []struct {
		in, want string
	}{
		{"Out of Service", "Out of Service"},           // exact member kept (canonical)
		{"out of service", "Out of Service"},           // case-insensitive -> canonical form
		{"  Standby / Spare ", "Standby / Spare"},      // trimmed, canonical
		{"decommissioned", "Decommissioned"},           // lowercase -> canonical
		{"Unknown weird thing", "New / Commissioning"}, // uncontrolled -> default
		{"", "New / Commissioning"},                    // empty -> default
	}
	for _, tc := range cases {
		out := ReconcileIdentity(IdentityInput{Status: tc.in}, nil, conv)
		if out.Status != tc.want {
			t.Errorf("status %q = %q, want %q", tc.in, out.Status, tc.want)
		}
	}
}

func TestReconcileIdentityStatusSynonymsCaseVariantKey(t *testing.T) {
	// A synonym-group key that is a case/trim variant of a declared status must
	// still match (and emit the declared canonical form) — not be skipped by an
	// exact-key contains check.
	conv := testOvaholConventions()
	conv.StatusSynonyms = map[string][]string{
		"in-service": {"functional", "working"},
	}
	out := ReconcileIdentity(IdentityInput{Status: "working"}, nil, conv)
	if out.Status != "In-Service" {
		t.Errorf("status %q = %q, want declared canonical %q", "working", out.Status, "In-Service")
	}
}

func TestReconcileIdentityStatusSynonyms(t *testing.T) {
	conv := testOvaholConventions()
	cases := []struct {
		in, want string
	}{
		// synonymous with In-Service
		{"functional", "In-Service"},
		{"FUNCTIONING", "In-Service"},
		{"Active", "In-Service"},
		{"in active service", "In-Service"},
		{"working", "In-Service"},
		// synonymous with Under Maintenance
		{"faulty", "Under Maintenance"},
		{"Not Working", "Under Maintenance"},
		{"Broken Down", "Under Maintenance"},
		{"broken down and repairable", "Under Maintenance"},
	}
	for _, tc := range cases {
		out := ReconcileIdentity(IdentityInput{Status: tc.in}, nil, conv)
		if out.Status != tc.want {
			t.Errorf("status %q = %q, want %q", tc.in, out.Status, tc.want)
		}
	}
}

func TestReconcileIdentityZeroValueConventions(t *testing.T) {
	// A zero-value Conventions must NOT invent strings: identity is left as-is.
	out := ReconcileIdentity(IdentityInput{Model: "m"}, nil, Conventions{})
	if out.Device != "" || out.Model != "m" || out.Brand != "" {
		t.Errorf("zero-value conventions altered identity: %+v", out)
	}
	if out.Status != "" {
		t.Errorf("zero-value conventions altered status: %q", out.Status)
	}
}

func TestEngineReconcile(t *testing.T) {
	engine := NewEngine(nil, WithConventions(testOvaholConventions()))
	out := engine.Reconcile(IdentityInput{DeviceName: "ECG machine", Status: "In-Service"})
	if out.Device != "ECG machine" || out.Model != "Unknown Model - ECG machine" {
		t.Errorf("engine.Reconcile: %+v", out)
	}
	if out.Status != "In-Service" {
		t.Errorf("engine.Reconcile status = %q, want In-Service", out.Status)
	}
}
