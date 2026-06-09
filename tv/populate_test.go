package tv

import "testing"

func TestPopulateSerieByLanguage_InsertPath(t *testing.T) {
	svc := NewService(nil, nil, nil)

	insertCalled := 0
	updateCalled := 0
	personCalls := 0

	svc.getSerieByIdLanguageFn = func(id int, language string) Serie { return Serie{Id: 0} }
	svc.insertSerieFn = func(itemObj Serie, language string) interface{} { insertCalled++; return nil }
	svc.updateSerieFn = func(itemObj Serie, language string) { updateCalled++ }
	svc.populatePersonFn = func(personId int, language, update string) { personCalls++ }

	serie := Serie{Id: 5, Title: "Show"}
	serie.TvCredits.Cast = []TvCast{{Id: 1}, {Id: 2}}
	serie.TvCredits.Crew = []TvCrew{{Id: 3}}

	svc.PopulateSerieByLanguage(serie, "en")

	if insertCalled != 1 || updateCalled != 0 {
		t.Fatalf("expected 1 insert, 0 update; got %d/%d", insertCalled, updateCalled)
	}
	if personCalls != 3 {
		t.Fatalf("expected 3 populatePerson calls, got %d", personCalls)
	}
}

func TestPopulateSerieByLanguage_UpdatePath(t *testing.T) {
	svc := NewService(nil, nil, nil)

	insertCalled := 0
	updateCalled := 0

	svc.getSerieByIdLanguageFn = func(id int, language string) Serie { return Serie{Id: 5} }
	svc.insertSerieFn = func(itemObj Serie, language string) interface{} { insertCalled++; return nil }
	svc.updateSerieFn = func(itemObj Serie, language string) { updateCalled++ }
	svc.populatePersonFn = func(personId int, language, update string) {
		t.Fatalf("populatePerson should not be called on update path")
	}

	svc.PopulateSerieByLanguage(Serie{Id: 5, Title: "Show"}, "en")

	if insertCalled != 0 || updateCalled != 1 {
		t.Fatalf("expected 0 insert, 1 update; got %d/%d", insertCalled, updateCalled)
	}
}

func TestPopulateSerieByLanguage_SkipPath(t *testing.T) {
	svc := NewService(nil, nil, nil)

	insertCalled := 0
	updateCalled := 0

	svc.getSerieByIdLanguageFn = func(id int, language string) Serie { return Serie{Id: 0} }
	svc.insertSerieFn = func(itemObj Serie, language string) interface{} { insertCalled++; return nil }
	svc.updateSerieFn = func(itemObj Serie, language string) { updateCalled++ }

	svc.PopulateSerieByLanguage(Serie{Id: 0}, "en")

	if insertCalled != 0 || updateCalled != 0 {
		t.Fatalf("expected no insert/update; got %d/%d", insertCalled, updateCalled)
	}
}
