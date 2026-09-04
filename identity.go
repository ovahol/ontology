package ontology

import "strings"

// Conventions defines how unresolved identity fields are handled and what
// controlled status values exist, for one vendor. Every string is supplied by
// the vendor — the engine hardcodes none of them. A zero-value Conventions
// performs no placement substitution and carries no status vocabulary,
// so identity resolution degrades gracefully to "leave it as-is".
type Conventions struct {
	// UnknownDevice is the placeholder used when no device name resolves.
	UnknownDevice string `json:"unknown_device,omitempty"`

	// UnknownBrand is the placeholder used when no brand resolves.
	UnknownBrand string `json:"unknown_brand,omitempty"`

	// UnknownManufacturer is the placeholder used when no manufacturer resolves.
	UnknownManufacturer string `json:"unknown_manufacturer,omitempty"`

	// UnknownModelPrefix prefixes the per-device unknown model placeholder.
	// The full placeholder is UnknownModelPrefix + device name, so each
	// unknown model stays unique per device (keeping the model->device
	// foreign key resolvable even for unknown models).
	UnknownModelPrefix string `json:"unknown_model_prefix,omitempty"`

	// Statuses is the vendor's controlled status vocabulary (e.g.
	// ["In-Service", "Out of Service"]). Order determines nothing; membership
	// controls which values are accepted verbatim.
	Statuses []string `json:"statuses,omitempty"`

	// StatusSynonyms maps any controlled status (a member of Statuses) to the
	// free-text terms a source inventory might use for it. A source status that
	// is neither a verbatim (case-insensitive) member of Statuses nor one of
	// these synonyms reconciles to DefaultStatus.
	StatusSynonyms map[string][]string `json:"status_synonyms,omitempty"`

	// DefaultStatus is used when the source supplies no status, an
	// uncontrolled one that is not in Statuses, or no matching synonym. When
	// present it should be a member of Statuses.
	DefaultStatus string `json:"default_status,omitempty"`
}

// UnknownModel returns the canonical placeholder model name for a device. It
// embeds the device name so that, even when the model is unknown, the
// model->device foreign key remains unique and resolvable.
func (c Conventions) UnknownModel(deviceName string) string {
	if c.UnknownModelPrefix == "" {
		return ""
	}
	if strings.TrimSpace(deviceName) == "" {
		return c.UnknownModelPrefix + c.UnknownDevice
	}
	return c.UnknownModelPrefix + deviceName
}

// IdentityInput is the raw, possibly-empty identity an inbound record carries:
// what the host system gave us before reconciliation. Every field is optional
// free-text.
type IdentityInput struct {
	DeviceName   string `json:"device_name,omitempty"`
	Model        string `json:"model_name,omitempty"`
	Brand        string `json:"brand_name,omitempty"`
	Manufacturer string `json:"manufacturer_name,omitempty"`
	Status       string `json:"status_name,omitempty"`
}

// IdentityResult is the reconciled identity view for one inbound record. After
// reconciliation every identity field is non-empty (Unknown placeholders fill
// gaps) and Status is a member of the vendor's controlled Statuses (or empty
// when the vendor models no status). The JSON keys mirror a lifecycle export.
type IdentityResult struct {
	Device       string `json:"device_name"`
	Model        string `json:"model_name"`
	Brand        string `json:"brand_name"`
	Manufacturer string `json:"manufacturer_name"`
	Status       string `json:"status_name"`
}

// IdentityResolver maps an inbound record's raw identity to the canonical
// device/model/brand/manufacturer in the host system's reference data. It is
// the mechanism by which a vendor's foreign-key chain (model->brand->
// manufacturer) is walked. Return empty strings for anything it cannot
// resolve; the engine fills those with the Unknown placeholders.
//
// A nil resolver is valid: everything resolves to its raw value and the
// engine applies the Unknown conventions only where said values are empty.
type IdentityResolver interface {
	// Resolve returns the canonical identity for an inbound record. The input
	// may be partially or fully empty. Return a zero (or empty) result when
	// nothing resolves.
	Resolve(in IdentityInput) (IdentityResult, bool)
}

// ReconcileIdentity produces a completed identity for an inbound record:
// it asks the resolver to canonicalize entity fields, falls back to the
// Unknown conventions for anything left unresolved/empty, and normalizes the
// status into the vendor's controlled vocabulary. A nil resolver behaves as
// "no entity resolution". A zero-value conv performs no substitution and no
// status normalization.
func ReconcileIdentity(in IdentityInput, res IdentityResolver, conv Conventions) IdentityResult {
	out := IdentityResult{
		Device:       strings.TrimSpace(in.DeviceName),
		Model:        strings.TrimSpace(in.Model),
		Brand:        strings.TrimSpace(in.Brand),
		Manufacturer: strings.TrimSpace(in.Manufacturer),
	}
	if res != nil {
		if r, ok := res.Resolve(in); ok {
			if strings.TrimSpace(r.Device) != "" {
				out.Device = strings.TrimSpace(r.Device)
			}
			if strings.TrimSpace(r.Model) != "" {
				out.Model = strings.TrimSpace(r.Model)
			}
			if strings.TrimSpace(r.Brand) != "" {
				out.Brand = strings.TrimSpace(r.Brand)
			}
			if strings.TrimSpace(r.Manufacturer) != "" {
				out.Manufacturer = strings.TrimSpace(r.Manufacturer)
			}
		}
	}

	// Apply Unknown conventions for whatever is still unresolved.
	if out.Device == "" {
		out.Device = strings.TrimSpace(conv.UnknownDevice)
	}
	if out.Model == "" {
		out.Model = conv.UnknownModel(out.Device)
	}
	if out.Brand == "" {
		out.Brand = strings.TrimSpace(conv.UnknownBrand)
	}
	if out.Manufacturer == "" {
		out.Manufacturer = strings.TrimSpace(conv.UnknownManufacturer)
	}

	out.Status = normalizeStatus(in.Status, conv.Statuses, conv.StatusSynonyms, conv.DefaultStatus)
	return out
}

// normalizeStatus maps a raw status into a controlled Statuses vocabulary.
// Matching is case-insensitive and trimmed. With no vocabulary (empty Statuses)
// the raw value is returned as-is. With a vocabulary: an exact member matches
// to its canonical form; a term listed under StatusSynonyms (for that canonical
// status) also matches to its canonical form; anything empty or unmatched maps
// to DefaultStatus (falling back to the empty string when none is set).
func normalizeStatus(raw string, statuses []string, synonyms map[string][]string, defaultStatus string) string {
	raw = strings.TrimSpace(raw)
	if len(statuses) == 0 {
		return raw
	}
	for _, s := range statuses {
		if strings.EqualFold(raw, strings.TrimSpace(s)) {
			return strings.TrimSpace(s)
		}
	}
	for canonical, terms := range synonyms {
		if !containsString(statuses, canonical) {
			continue
		}
		for _, term := range terms {
			if strings.EqualFold(raw, strings.TrimSpace(term)) {
				return canonical
			}
		}
	}
	return strings.TrimSpace(defaultStatus)
}

func containsString(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}
