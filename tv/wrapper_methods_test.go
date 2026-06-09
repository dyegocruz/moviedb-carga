package tv

import (
	"reflect"
	"testing"

	"moviedb/common"
)

func TestTvWrappers_UseHooks(t *testing.T) {
	svc := NewService(nil, nil, nil)

	svc.getSerieByIdLanguageFn = func(id int, language string) Serie {
		return Serie{Id: id, Language: language, Title: "hooked-serie"}
	}
	svc.insertSerieFn = func(itemObj Serie, language string) interface{} { return "serie-id" }
	updatedSerie := 0
	svc.updateSerieFn = func(itemObj Serie, language string) { updatedSerie++ }
	deletedSerie := 0
	svc.deleteSerieFn = func(id int) { deletedSerie = id }

	svc.insertEpisodeFn = func(itemObj Episode, language string) interface{} { return "episode-id" }
	svc.getEpisodeByIdLanguageFn = func(id int, language string) Episode {
		return Episode{Id: id, Language: language}
	}
	svc.getEpisodeBySerieSeasonLanguageFn = func(showId int, seasonNumber int, language string) []Episode {
		return []Episode{{Id: 11}, {Id: 22}}
	}
	updatedEpisode := 0
	svc.updateEpisodeFn = func(itemObj Episode, language string) { updatedEpisode++ }
	deletedEpisodes := 0
	svc.deleteSerieEpisodesFn = func(showId int) { deletedEpisodes = showId }

	svc.getCountAllFn = func() int64 { return 123 }
	svc.generateTvCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{1: {Id: 1}}
	}
	svc.generateTvEpisodesCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{2: {Id: 2}}
	}

	serie := svc.GetSerieByIdAndLanguage(9, "en")
	if serie.Id != 9 || serie.Title != "hooked-serie" {
		t.Fatalf("unexpected serie from hook: %+v", serie)
	}
	if got := svc.InsertSerie(Serie{Id: 9}, "en"); got != "serie-id" {
		t.Fatalf("unexpected insert serie return: %v", got)
	}
	svc.UpdateSerie(Serie{Id: 9}, "en")
	if updatedSerie != 1 {
		t.Fatalf("expected update serie hook call, got %d", updatedSerie)
	}
	svc.DeleteSerie(55)
	if deletedSerie != 55 {
		t.Fatalf("expected delete serie id 55, got %d", deletedSerie)
	}

	if got := svc.InsertEpisode(Episode{Id: 77}, "pt-BR"); got != "episode-id" {
		t.Fatalf("unexpected insert episode return: %v", got)
	}
	ep := svc.GetEpisodeByIdAndLanguage(77, "pt-BR")
	if ep.Id != 77 || ep.Language != "pt-BR" {
		t.Fatalf("unexpected episode from hook: %+v", ep)
	}
	episodes := svc.GetEpisodeBySerieSeasonAndLanguage(1, 2, "en")
	if !reflect.DeepEqual(episodes, []Episode{{Id: 11}, {Id: 22}}) {
		t.Fatalf("unexpected episodes: %+v", episodes)
	}
	svc.UpdateEpisode(Episode{Id: 10}, "en")
	if updatedEpisode != 1 {
		t.Fatalf("expected update episode hook call, got %d", updatedEpisode)
	}
	svc.DeleteSerieEpisodes(66)
	if deletedEpisodes != 66 {
		t.Fatalf("expected delete episodes showId 66, got %d", deletedEpisodes)
	}

	if got := svc.GetCountAll(); got != 123 {
		t.Fatalf("unexpected count: %d", got)
	}
	if got := svc.GenerateTvCatalogCheck("en"); got[1].Id != 1 {
		t.Fatalf("unexpected tv catalog check: %+v", got)
	}
	if got := svc.GenerateTvEpisodesCatalogCheck("en"); got[2].Id != 2 {
		t.Fatalf("unexpected tv episodes catalog check: %+v", got)
	}
}
