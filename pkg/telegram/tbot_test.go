package telegram

import "testing"

func TestParseMessage(t *testing.T) {
	t.Run("good commands", func(t *testing.T) {
		list := []string{
			"/add",
			"/add naruto",
			"/add one piece",
			"/add  ",
			"/add the fragrant flower blooms with dignity ",
		}
		expected := []string{
			"",
			"naruto",
			"one piece",
			"",
			"the fragrant flower blooms with dignity",
		}

		for i, msg := range list {
			parsed, err := parseMessage("/add", msg)
			if err != nil {
				t.Errorf("error in parsing command: %s\n", msg)
			}
			if parsed != expected[i] {
				t.Errorf("For input [%s],\nexpected [%s],\ngot [%s]\n", msg, expected[i], parsed)
			}
		}
	})

	t.Run("bad commands", func(t *testing.T) {
		badList := []string{
			"hello /add",
			"hello new world",
			"",
		}

		for _, msg := range badList {
			_, err := parseMessage("/add", msg)
			if err == nil {
				t.Errorf("expected error for bad command: [%s], got nil\n", msg)
			}
		}
	})
}
