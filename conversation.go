package main

import "fmt"

type ChatID int64
type AddMangaConversationState int
type CommandManga string


const (
	StartConversation AddMangaConversationState = iota + 1
	ChosenManga
	ChoseWhatToDo
)

const (
	Download 	CommandManga = "Download"
	ReadOnline	CommandManga = "Read Online"
	DoNothing  	CommandManga = "Do Nothing"
)


// ConversationsStore is used to save the choices of the user during a conversation with the bot.
// the encapsulated maps are not thread safe, but it is not needed now because of very low usage
// when the user is done with the conversation, it is important to call the Candel method
type ConversationsStore struct {
	addManga	 map[ChatID]AddMangaConversationState
	commandManga map[ChatID]CommandManga
	mangas		 map[ChatID][]Manga
	chosenManga	 map[ChatID]Manga
}

func NewConversationStore() ConversationsStore {
	return ConversationsStore{
		addManga: 		make(map[ChatID]AddMangaConversationState),
		commandManga:	make(map[ChatID]CommandManga),
		mangas: 		make(map[ChatID][]Manga),
		chosenManga: 	make(map[ChatID]Manga),
	}
}

func (s ConversationsStore) InsertAddMangaState(chatID ChatID, state AddMangaConversationState) {
	// i think for now its irrelevant to check if the chatID has already a state
	s.addManga[chatID] = state 
}

func (s ConversationsStore) InsertCommandManga(chatID ChatID, command CommandManga) {
	s.commandManga[chatID] = command
}

func (s ConversationsStore) InsertMangas(chatID ChatID, mangas []Manga) {
	s.mangas[chatID] = mangas
}

func (s ConversationsStore) InsertChosenManga(chatID ChatID, manga Manga) {
	s.chosenManga[chatID] = manga
}

func (s ConversationsStore) GetAddMangaState(chatID ChatID) (AddMangaConversationState, error) {
	state, ok := s.addManga[chatID]
	if !ok {
		return 0, fmt.Errorf("addManga state not found for chatID %v", chatID)
	}
	return state, nil
}

func (s ConversationsStore) GetCommandManga(chatID ChatID) (CommandManga, error) {
	command, ok := s.commandManga[chatID]
	if !ok {
		return "", fmt.Errorf("commandManga not found for chatID %v", chatID)
	}
	return command, nil
}

func (s ConversationsStore) GetMangas(chatID ChatID) ([]Manga, error) {
	mangas, ok := s.mangas[chatID]
	if !ok {
		return nil, fmt.Errorf("mangas not found for chatID %v", chatID)
	}
	return mangas, nil
}

func (s ConversationsStore) GetChosenManga(chatID ChatID) (Manga, error) {
	manga, ok := s.chosenManga[chatID]
	if !ok {
		return Manga{}, fmt.Errorf("chosenManga not found for chatID %v", chatID)
	}
	return manga, nil
}

func (s ConversationsStore) Clean(chatID ChatID) {
	delete(s.addManga, chatID)
	delete(s.commandManga, chatID)
	delete(s.mangas, chatID)
	delete(s.chosenManga, chatID)
}