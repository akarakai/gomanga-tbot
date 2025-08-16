package repository

import (
	"database/sql"

	"github.com/akarakai/gomanga-tbot/pkg/logger"
	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type UserRepo interface {
	SaveUser(chatID model.ChatID) error
	SaveManga(chatID model.ChatID, mangaUrl string) error
	FindUserByChatID(chatID model.ChatID) (*model.User, error)
	FindAllUsers() ([]model.User, error)
}

type UserRepoSqlite3 struct {
	db *sql.DB
}

func (repo *UserRepoSqlite3) SaveUser(chatID model.ChatID) error {
	_, err := repo.db.Exec(`
		INSERT OR IGNORE INTO users (chat_id)
		VALUES (?)
	`, chatID)
	if err != nil {
		logger.Log.Errorw("error when saving user", "chat_id", chatID, "err", err)
		return err
	}

	logger.Log.Debugw("user saved successfully", "chat_id", chatID)
	return nil
}

func (repo *UserRepoSqlite3) FindUserByChatID(chatID model.ChatID) (*model.User, error) {
	row := repo.db.QueryRow(`
        SELECT chat_id FROM users
        WHERE chat_id = ?
    `, chatID)

	var chatIDq sql.NullInt64
	if err := row.Scan(&chatIDq); err != nil {
		if err == sql.ErrNoRows {
			logger.Log.Debugw("user does not exist", "chat_id", chatIDq.Int64)
			return nil, nil
		}
		logger.Log.Errorw("error when scanning user row", "err", err)
		return nil, err
	}

	logger.Log.Debugw("user found successfully", "chat_id", chatIDq.Int64)

	return &model.User{
		ChatID: model.ChatID(chatIDq.Int64),
	}, nil
}

func (repo *UserRepoSqlite3) SaveManga(chatID model.ChatID, mangaUrl string) error {
	_, err := repo.db.Exec(`
		INSERT INTO user_mangas (chat_id, manga_url)
		VALUES (?, ?)
	`, chatID, mangaUrl)
	if err != nil {
		logger.Log.Errorw("error when saving manga row", "err", err)
		return err
	}
	logger.Log.Debugw("manga saved successfully in user", "chat_id", chatID)
	return nil
}

// finds also the mangas of a user in order to complete the User struct and the chapter of each 
func (repo *UserRepoSqlite3) FindAllUsers() ([]model.User, error) {
	rows, err := repo.db.Query(`
		SELECT
			u.chat_id,
			m.url       AS manga_url,
			m.title     AS manga_title,
			c.url       AS chapter_url,
			c.title     AS chapter_title,
			c.released_at
		FROM users u
		LEFT JOIN user_mangas um ON um.chat_id = u.chat_id
		LEFT JOIN mangas m       ON m.url      = um.manga_url
		LEFT JOIN chapters c     ON c.url      = m.last_chapter
	`)
	if err != nil {
		logger.Log.Errorw("FindAllUsers: query failed", "err", err)
		return nil, err
	}
	defer func() {
		if cerr := rows.Close(); cerr != nil {
			logger.Log.Errorw("FindAllUsers: close rows failed", "err", cerr)
		}
	}()

	usersByID := make(map[model.ChatID]*model.User)

	for rows.Next() {
		var (
			chatID                 model.ChatID
			mangaURL, mangaTitle   sql.NullString
			chURL, chTitle         sql.NullString
			chReleased             sql.NullTime
		)

		if err := rows.Scan(
			&chatID,
			&mangaURL, &mangaTitle,
			&chURL, &chTitle, &chReleased,
		); err != nil {
			logger.Log.Errorw("FindAllUsers: scan failed", "err", err)
			return nil, err
		}

		key := model.ChatID(chatID)
		u, ok := usersByID[key]
		if !ok {
			u = &model.User{
				ChatID: key,
				// Mangas will be appended below if present
			}
			usersByID[key] = u
		}

		// If user has at least one manga row
		if mangaURL.Valid {
			m := model.Manga{
				Url:   mangaURL.String,
				Title: mangaTitle.String,
			}

			if chURL.Valid {
				ch := &model.Chapter{
					Url:   chURL.String,
					Title: chTitle.String,
				}
				if chReleased.Valid {
					ch.ReleasedAt = chReleased.Time
				}
				m.LastChapter = ch
			}

			u.Mangas = append(u.Mangas, m)
		}
	}

	if err := rows.Err(); err != nil {
		logger.Log.Errorw("FindAllUsers: row iter error", "err", err)
		return nil, err
	}

	// Flatten map -> slice
	out := make([]model.User, 0, len(usersByID))
	for _, u := range usersByID {
		out = append(out, *u)
	}
	return out, nil
}
