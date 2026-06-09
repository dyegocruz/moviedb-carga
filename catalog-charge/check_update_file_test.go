package catalogCharge

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"moviedb/common"
	"moviedb/queue"
	"moviedb/services"
)

type fakeRabbitPublisher struct {
	closed    bool
	published []queue.CatalogProcessMessage
}

func (f *fakeRabbitPublisher) Close() { f.closed = true }

func (f *fakeRabbitPublisher) PublishJSON(queueName string, data interface{}) error {
	msg, ok := data.(queue.CatalogProcessMessage)
	if ok {
		f.published = append(f.published, msg)
	}
	return nil
}

func TestCheckAndUpdateCatalogByFile_MoviePublishesAndDeletesMissing(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "daily.json")
	content := "{\"id\":1}\n{\"id\":3}\n"
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	rabbit := &fakeRabbitPublisher{}
	deleted := make([]int, 0)
	var removedFile string
	var downloaded, unzipped string

	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.nowFn = func() time.Time { return time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC) }
	svc.downloadExportFileFn = func(url string, fileName string) { downloaded = fileName }
	svc.unzipFn = func(fileName string) { unzipped = fileName }
	svc.openFileFn = func(name string) (*os.File, error) { return os.Open(filePath) }
	svc.newRabbitMQServiceFn = func(config services.Config) (rabbitPublisher, error) {
		return rabbit, nil
	}
	svc.removeFileFn = func(name string) { removedFile = name }
	svc.generateMovieCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{1: {Id: 1}, 2: {Id: 2}}
	}
	svc.deleteMovieFn = func(id int) { deleted = append(deleted, id) }

	svc.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)

	if !rabbit.closed {
		t.Fatal("expected rabbit to be closed")
	}
	if !reflect.DeepEqual(rabbit.published, []queue.CatalogProcessMessage{{Id: 3, MediaType: common.MEDIA_TYPE_MOVIE}}) {
		t.Fatalf("unexpected published messages: %+v", rabbit.published)
	}
	if !reflect.DeepEqual(deleted, []int{2}) {
		t.Fatalf("unexpected deleted ids: %+v", deleted)
	}
	if downloaded != "movie_ids_06_02_2026" || unzipped != "movie_ids_06_02_2026" {
		t.Fatalf("unexpected download/unzip fileName: %s / %s", downloaded, unzipped)
	}
	if removedFile != "movie_ids_06_02_2026.json" {
		t.Fatalf("unexpected removed file: %s", removedFile)
	}
}

func TestCheckAndUpdateCatalogByFile_TvDeletesSerieAndEpisodes(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "daily_tv.json")
	content := "{\"id\":10}\n{\"id\":30}\n"
	if err := os.WriteFile(filePath, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	rabbit := &fakeRabbitPublisher{}
	deletedSeries := make([]int, 0)
	deletedEpisodes := make([]int, 0)

	svc := NewService(nil, nil, nil, nil, nil, nil)
	svc.nowFn = func() time.Time { return time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC) }
	svc.downloadExportFileFn = func(url string, fileName string) {}
	svc.unzipFn = func(fileName string) {}
	svc.openFileFn = func(name string) (*os.File, error) { return os.Open(filePath) }
	svc.newRabbitMQServiceFn = func(config services.Config) (rabbitPublisher, error) {
		return rabbit, nil
	}
	svc.removeFileFn = func(name string) {}
	svc.generateTvCatalogCheckFn = func(language string) map[int]common.CatalogCheck {
		return map[int]common.CatalogCheck{10: {Id: 10}, 20: {Id: 20}}
	}
	svc.deleteSerieFn = func(id int) { deletedSeries = append(deletedSeries, id) }
	svc.deleteSerieEpisodesFn = func(id int) { deletedEpisodes = append(deletedEpisodes, id) }

	svc.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)

	if !rabbit.closed {
		t.Fatal("expected rabbit to be closed")
	}
	if !reflect.DeepEqual(rabbit.published, []queue.CatalogProcessMessage{{Id: 30, MediaType: common.MEDIA_TYPE_TV}}) {
		t.Fatalf("unexpected published messages: %+v", rabbit.published)
	}
	if !reflect.DeepEqual(deletedSeries, []int{20}) {
		t.Fatalf("unexpected deleted series ids: %+v", deletedSeries)
	}
	if !reflect.DeepEqual(deletedEpisodes, []int{20}) {
		t.Fatalf("unexpected deleted episode ids: %+v", deletedEpisodes)
	}
}
