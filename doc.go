// Package ontology is a system-agnostic device interchange and classification
// engine. It executes vendor-defined vocabularies, taxonomies, and inference
// rules. ontology does not own a vendor's vocabulary — each system supplies
// its own taxonomy, and ontology acts purely as the engine that normalizes and
// classifies device records against it.
//
// # What a vendor provides
//
// A vendor defines their controlled vocabulary and rules in their own codebase,
// expressed as a Taxonomy (JSON). A Taxonomy is just two things: Fields, the
// list of classification dimensions the vendor cares about (each with its own
// key and controlled vocabulary — there's no fixed set of dimensions, and no
// dimension name is special to the engine), and Inference.Rules, an ordered
// list of keyword/source-type/field-dependency conditions that assign values
// to those fields. Load it with LoadTaxonomyFile.
//
// If no taxonomy is supplied, ontology falls back to an embedded default
// derived from the WHO's MeDevIS reference vocabulary (DefaultTaxonomy). This
// default exists only so nil-taxonomy callers still produce useful,
// standards-aligned output — it is not a vendor's vocabulary.
//
// # Quick start
//
//	import "github.com/ovahol/ontology"
//
//	tax, err := ontology.LoadTaxonomyFile("my_taxonomy.json")
//	result := ontology.NormalizeWithTaxonomy(ontology.Input{
//	    DeviceName: "ECG machine, portable, 12-lead",
//	    SourceType: "monitoring equipment",
//	    EMDNTerm:   "Electrocardiographs",
//	}, tax)
//
// Or build an Engine bound to a taxonomy and/or catalog:
//
//	engine := ontology.NewEngine(tax, ontology.WithCatalog(myCatalog))
//	result := engine.Normalize(input)
//
// # Identity reconciliation
//
// Classification tells you what a device is; reconciliation tells you how it
// exists in the vendor's system — its name, model, brand, manufacturer, and a
// controlled status. The engine provides an identity-reconciliation mechanism
// that is also entirely vendor-configured, so no identity vocabulary is
// hardcoded here:
//
//	engine := ontology.NewEngine(tax,
//	    ontology.WithCatalog(myCatalog),
//	    ontology.WithConventions(myConventions),        // Unknown placeholders + status set
//	    ontology.WithIdentityResolver(myResolver),      // walks model->brand->manufacturer
//	)
//	id := engine.Reconcile(ontology.IdentityInput{
//	    DeviceName: result.Name, // the classified name
//	    Model:      src.Model,
//	    Brand:      src.Brand,
//	    Manufacturer: src.Manufacturer,
//	    Status:     src.Status,
//	})
//
// `Conventions` is the vendor's Unknown placeholder templates, its controlled
// `Statuses` vocabulary (with a `DefaultStatus`), and `StatusSynonyms` mapping
// free-text source statuses (e.g. "functional", "broken down") to a canonical
// status. `IdentityResolver`
// maps an inbound identity tuple to the canonical entity names via the vendor's
// own reference-data foreign keys; return empty strings for anything you cannot
// resolve and the engine fills the Unknown placeholders. ReconcileIdentity is
// the resolver-free counterpart.
//
// # Public entry points
//
//   - NormalizeWithTaxonomy / NormalizeBatchWithTaxonomy: single/batch records
//   - NormalizeWithCatalogAndTaxonomy: catalog-first exact match, taxonomy fallback
//   - NormalizeWorkbookWithTaxonomy / NormalizeCSVWithTaxonomy: file bulk import
//   - NormalizeJSONWithTaxonomy: JSON arrays/objects
//   - ReconcileIdentity: completed identity (Unknown placeholders + status)
//   - Engine: NewEngine + WithTaxonomy/WithCatalog/WithConventions/WithIdentityResolver
//   - DefaultTaxonomy: the embedded WHO/MeDevIS reference vocabulary
//
// # Interchange schema
//
// The library defines a language-agnostic JSON interchange schema so
// non-Go systems can participate by producing/consuming JSON:
//
//	{
//	  "device_name": "Infusion pump, volumetric",
//	  "source_type": "infusion devices",
//	  "emdn_code": "Z1201",
//	  "emdn_term": "Volumetric infusion pumps"
//	}
//
// Normalizes to:
//
//	{
//	  "name": "Infusion pump",
//	  "fields": {
//	    "device_type": "Treatment, Surgical & Life Support Devices",
//	    "device_category": "Therapeutic",
//	    "device_function": "Surgical and Intensive Care",
//	    "device_application_risk": "Potential patient or operator injury"
//	  },
//	  "mapping_source": "family_fallback"
//	}
//
// The controlled vocabulary is entirely determined by the vendor taxonomy —
// no dimension (device_type, service_type, knowledge_level, reusable, or any
// vendor-defined key) is special to the engine, and no free-text leaks into
// the controlled fields. The embedded default taxonomy derives its own
// dimensions from the WHO MeDevIS reference (device_type, service_type,
// knowledge_level, reusable, plus EMDN/GMDN lookups); a vendor like Ovahol
// declares its own instead.
package ontology
