package telegram

import (
	"fmt"

	"github.com/akarakai/gomanga-tbot/pkg/model"
)

type AddMangaConversationState int
type CommandManga string

const (
	StartConversation AddMangaConversationState = iota + 1
	ChosenManga
	ChoseWhatToDo
)

const (
	Download   CommandManga = "Download"
	ReadOnline CommandManga = "Read Online"
	DoNothing  CommandManga = "Do Nothing"
)

var convStore = NewConversationStore()

// ConversationsStore is used to save the choices of the user during a conversation with the bot.
// the encapsulated maps are not thread safe, but it is not needed now because of very low usage
// when the user is done with the conversation, it is important to call the Candel method
type ConversationsStore struct {
	addManga     map[model.ChatID]AddMangaConversationState
	commandManga map[model.ChatID]CommandManga
	mangas       map[model.ChatID][]model.Manga
	chosenManga  map[model.ChatID]model.Manga
}

func NewConversationStore() ConversationsStore {
	return ConversationsStore{
		addManga:     make(map[model.ChatID]AddMangaConversationState),
		commandManga: make(map[model.ChatID]CommandManga),
		mangas:       make(map[model.ChatID][]model.Manga),
		chosenManga:  make(map[model.ChatID]model.Manga),
	}
}

func (s ConversationsStore) InsertAddMangaState(chatID model.ChatID, state AddMangaConversationState) {
	// i think for now its irrelevant to check if the chatID has already a state
	s.addManga[chatID] = state
}

func (s ConversationsStore) InsertCommandManga(chatID model.ChatID, command CommandManga) {
	s.commandManga[chatID] = command
}

func (s ConversationsStore) InsertMangas(chatID model.ChatID, mangas []model.Manga) {
	s.mangas[chatID] = mangas
}

func (s ConversationsStore) InsertChosenManga(chatID model.ChatID, manga model.Manga) {
	s.chosenManga[chatID] = manga
}

func (s ConversationsStore) GetAddMangaState(chatID model.ChatID) (AddMangaConversationState, error) {
	state, ok := s.addManga[chatID]
	if !ok {
		return 0, fmt.Errorf("addManga state not found for chatID %v", chatID)
	}
	return state, nil
}

func (s ConversationsStore) GetCommandManga(chatID model.ChatID) (CommandManga, error) {
	command, ok := s.commandManga[chatID]
	if !ok {
		return "", fmt.Errorf("commandManga not found for chatID %v", chatID)
	}
	return command, nil
}

func (s ConversationsStore) GetMangas(chatID model.ChatID) ([]model.Manga, error) {
	mangas, ok := s.mangas[chatID]
	if !ok {
		return nil, fmt.Errorf("mangas not found for chatID %v", chatID)
	}
	return mangas, nil
}

func (s ConversationsStore) GetChosenManga(chatID model.ChatID) (model.Manga, error) {
	manga, ok := s.chosenManga[chatID]
	if !ok {
		return model.Manga{}, fmt.Errorf("chosenManga not found for chatID %v", chatID)
	}
	return manga, nil
}

func (s ConversationsStore) Clean(chatID model.ChatID) {
	delete(s.addManga, chatID)
	delete(s.commandManga, chatID)
	delete(s.mangas, chatID)
	delete(s.chosenManga, chatID)
}
