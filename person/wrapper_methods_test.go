package person

import (
	"reflect"
	"sync"
	"testing"

	"moviedb/common"
)

func TestPersonWrappers_UseHooks(t *testing.T) {
	svc := NewService(nil, nil)

	svc.getAllByIdsFn = func(ids []int) []interface{} {
		return []interface{}{Person{Id: ids[0], Name: "P1"}}
	}
	svc.getCatalogSearchInFn = func(language string, ids []int) []Person {
		return []Person{{Id: ids[0], Language: language, Name: "Catalog"}}
	}
	svc.getPersonByIdLanguageFn = func(id int, language string) Person {
		return Person{Id: id, Language: language, Name: "ByID"}
	}
	svc.getPersonsWithCreditsFn = func(language string) []Person {
		return []Person{{Id: 10, Language: language}}
	}
	svc.getPersonsWithoutCreditsFn = func(language string) []Person {
		return []Person{{Id: 20, Language: language}}
	}

	svc.insertPersonFn = func(itemObj Person) interface{} { return "inserted-id" }
	updated := 0
	svc.updatePersonFn = func(itemObj Person, language string) { updated++ }

	svc.getCountAllFn = func() int64 { return 99 }
	svc.generatePersonCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{1: {Id: 1}}
	}

	allByIds := svc.GetAllByIds([]int{7})
	if len(allByIds) != 1 {
		t.Fatalf("expected one doc, got %d", len(allByIds))
	}

	catalogIn := svc.GetCatalogSearchIn("en", []int{8})
	if len(catalogIn) != 1 || catalogIn[0].Id != 8 || catalogIn[0].Language != "en" {
		t.Fatalf("unexpected catalog docs: %+v", catalogIn)
	}

	personByID := svc.GetPersonByIdAndLanguage(11, "pt-BR")
	if personByID.Id != 11 || personByID.Language != "pt-BR" {
		t.Fatalf("unexpected person by id+language: %+v", personByID)
	}

	withCredits := svc.GetPersonsWithCredits("en")
	if !reflect.DeepEqual(withCredits, []Person{{Id: 10, Language: "en"}}) {
		t.Fatalf("unexpected with credits: %+v", withCredits)
	}

	withoutCredits := svc.GetPersonsWithoutCredits("pt-BR")
	if !reflect.DeepEqual(withoutCredits, []Person{{Id: 20, Language: "pt-BR"}}) {
		t.Fatalf("unexpected without credits: %+v", withoutCredits)
	}

	if inserted := svc.InsertPerson(Person{Id: 12}); inserted != "inserted-id" {
		t.Fatalf("unexpected inserted id: %v", inserted)
	}

	svc.UpdatePerson(Person{Id: 13}, "en")
	if updated != 1 {
		t.Fatalf("expected update hook once, got %d", updated)
	}

	if count := svc.GetCountAll(); count != 99 {
		t.Fatalf("unexpected count: %d", count)
	}

	catalogCheck := svc.GeneratePersonCatalogCheck("en")
	if catalogCheck[1].Id != 1 {
		t.Fatalf("unexpected catalog check: %+v", catalogCheck)
	}
}

func TestPopulatePersons_UsesHooks(t *testing.T) {
	svc := NewService(nil, nil)

	svc.maxPageLoadFn = func() int { return 2 }
	svc.getPopularPersonFn = func(language string, page string) ResultPerson {
		if page == "1" {
			return ResultPerson{Results: []Person{{Id: 1}, {Id: 0}}}
		}
		return ResultPerson{Results: []Person{{Id: 2}}}
	}
	svc.getPersonDetailsFn = func(id int, language string) Person {
		return Person{Id: id, Name: "P", Language: language}
	}

	var mu sync.Mutex
	calls := map[string]int{}
	done := make(chan struct{}, 4)
	svc.populatePersonByLanguageFn = func(itemObj Person, language string, updatePerson string) {
		mu.Lock()
		calls[language]++
		mu.Unlock()
		done <- struct{}{}
	}

	svc.PopulatePersons("en")

	for i := 0; i < 4; i++ {
		<-done
	}

	mu.Lock()
	defer mu.Unlock()
	if calls[common.LANGUAGE_PTBR] != 2 {
		t.Fatalf("expected 2 pt-BR calls, got %d", calls[common.LANGUAGE_PTBR])
	}
	if calls[common.LANGUAGE_EN] != 2 {
		t.Fatalf("expected 2 en calls, got %d", calls[common.LANGUAGE_EN])
	}
}
