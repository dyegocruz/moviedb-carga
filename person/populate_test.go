package person

import "testing"

func TestPopulatePersonByLanguage_InsertPath(t *testing.T) {
	svc := NewService(nil, nil)

	insertCalled := 0
	updateCalled := 0

	svc.getPersonByIdLanguageFn = func(id int, language string) Person { return Person{Id: 0} }
	svc.insertPersonFn = func(itemObj Person) interface{} { insertCalled++; return nil }
	svc.updatePersonFn = func(itemObj Person, language string) { updateCalled++ }

	svc.PopulatePersonByLanguage(Person{Id: 7, Name: "Alice"}, "en", "N")

	if insertCalled != 1 || updateCalled != 0 {
		t.Fatalf("expected 1 insert, 0 update; got %d/%d", insertCalled, updateCalled)
	}
}

func TestPopulatePersonByLanguage_UpdatePath(t *testing.T) {
	svc := NewService(nil, nil)

	insertCalled := 0
	updateCalled := 0

	svc.getPersonByIdLanguageFn = func(id int, language string) Person { return Person{Id: 7} }
	svc.insertPersonFn = func(itemObj Person) interface{} { insertCalled++; return nil }
	svc.updatePersonFn = func(itemObj Person, language string) { updateCalled++ }

	svc.PopulatePersonByLanguage(Person{Id: 7, Name: "Alice"}, "en", "Y")

	if insertCalled != 0 || updateCalled != 1 {
		t.Fatalf("expected 0 insert, 1 update; got %d/%d", insertCalled, updateCalled)
	}
}

func TestPopulatePersonByLanguage_SkipPath(t *testing.T) {
	svc := NewService(nil, nil)

	insertCalled := 0
	updateCalled := 0

	// existing != 0 but updatePerson != "Y" -> skip
	svc.getPersonByIdLanguageFn = func(id int, language string) Person { return Person{Id: 7} }
	svc.insertPersonFn = func(itemObj Person) interface{} { insertCalled++; return nil }
	svc.updatePersonFn = func(itemObj Person, language string) { updateCalled++ }

	svc.PopulatePersonByLanguage(Person{Id: 7}, "en", "N")

	if insertCalled != 0 || updateCalled != 0 {
		t.Fatalf("expected no insert/update; got %d/%d", insertCalled, updateCalled)
	}
}
