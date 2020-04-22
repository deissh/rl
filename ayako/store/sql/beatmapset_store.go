package sql

import (
	"github.com/deissh/osu-lazer/ayako/entity"
	"github.com/deissh/osu-lazer/ayako/store"
	"github.com/rs/zerolog/log"
)

type BeatmapSetStore struct {
	SqlStore
}

func newSqlBeatmapSetStore(sqlStore SqlStore) store.BeatmapSet {
	return &BeatmapSetStore{sqlStore}
}

func (s BeatmapSetStore) GetBeatmapSet(id uint) (*entity.BeatmapSetFull, error) {
	var set entity.BeatmapSetFull

	err := s.GetMaster().Get(
		&set,
		`SELECT id, title, artist, play_count, favourite_count,
			has_favourited, submitted_date, last_updated, ranked_date,
		   creator, user_id, bpm, source, covers, preview_url, tags, video,
		   storyboard, ranked, status, is_scoreable, discussion_enabled,
		   discussion_locked, can_be_hyped, availability, hype, nominations,
		   legacy_thread_url, description, genre, language, "user"
		FROM beatmap_set
		WHERE id = $1;`,
		id,
	)
	if err != nil {
		log.Error().
			Err(err).
			Msg("store.GetBeatmapSet")

		//todo: error wrap
		return nil, err
	}

	return &set, nil
}

func (s BeatmapSetStore) GetAllBeatmapSets(page int, limit int) (*[]entity.BeatmapSet, error) {
	panic("implement me")
}

func (s BeatmapSetStore) CreateBeatmapSet(from interface{}) (*entity.BeatmapSetFull, error) {
	panic("implement me")
}

func (s BeatmapSetStore) UpdateBeatmapSet(id uint, from interface{}) (*entity.BeatmapSetFull, error) {
	panic("implement me")
}

func (s BeatmapSetStore) DeleteBeatmapSet(id uint) error {
	panic("implement me")
}

func (s BeatmapSetStore) Fetch(id uint, merge bool) (*entity.BeatmapSetFull, error) {
	data, err := s.GetOsuClient().BeatmapSet.Get(id)
	if err != nil {
		return nil, err
	}

	log.Debug().
		Int64("id", data.ID).
		Str("name", data.Title).
		Str("updated_at", data.LastUpdated.String()).
		Send()

	return nil, nil
}
