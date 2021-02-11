package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	ui "github.com/gizak/termui"
	"github.com/gizak/termui/widgets"
)

type EntryList struct {
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Date    string `json:"date"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Id      string `json:"id"`
}

type View string

const (
	Editor    View = "editor"
	Browser        = "browser"
	EditTitle      = "edit_title"
)

var get_entries_endpoint = "get_entry.php"
var add_entry_endpoint = "add_entry.php"
var delete_entry_endpoint = "delete_entry.php"
var modify_entry_endpoint = "modify_entry.php"

var view View = Browser
var header = widgets.NewParagraph()
var notes = widgets.NewList()
var grid = ui.NewGrid()
var content = widgets.NewParagraph()

var entry_list EntryList
var caret_pos int
var raw_text string
var base_url string

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func edit_text(input string, to_modify *string) {
	switch input {
	case "<Left>":
		caret_pos = max(0, caret_pos-1.0)
	case "<Right>":
		caret_pos = min(len(raw_text), caret_pos+1)
	case "<Space>", "<Enter>", "0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "!", "\"", "#", "$", "%", "&", "'", "(", ")", "*", "+", ",", "-", ".", "/", ":", ";", "<", "=", ">", "?", "@", "[", "\\", "]", "^", "_", "`", "{", "|", "}", "~":
		if input == "<Space>" {
			input = " "
		} else if input == "<Enter>" {
			input = "\n"
		}

		raw_text = raw_text[:caret_pos] + input + raw_text[caret_pos:]
		caret_pos++
	case "<Backspace>":
		raw_text = raw_text[:max(0, caret_pos-1.0)] + raw_text[caret_pos:]
		caret_pos--
	case "<Delete>":
		if caret_pos != len(raw_text) {
			raw_text = raw_text[:caret_pos] + raw_text[min(len(raw_text), caret_pos+1.0):]
		}
	}

	caret_pos = max(0, min(len(raw_text)-1, caret_pos))

	if len(raw_text) == 0 || raw_text[len(raw_text)-1] != ' ' {
		raw_text = raw_text + " "
	}

	highlighted_char := raw_text[caret_pos]

	if highlighted_char == ' ' {
		highlighted_char = '_'
	}

	if view == Editor {
		*to_modify = raw_text[:caret_pos] + "[" + string(highlighted_char) + "](fg:green,mod:bold)" + raw_text[caret_pos+1:]
	} else if view == EditTitle {
		*to_modify = raw_text[:caret_pos] + "|" + raw_text[caret_pos:]
	}
}

func show_error(err string) {
	fmt.Println("Error: " + err)
	os.Exit(-1)
}

func db_query(endpoint string) string {
	resp, err := http.Get(base_url + endpoint)

	if err != nil {
		show_error("Can't connect to database.")
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		show_error("Can't read response body.")
	} else if resp.StatusCode != 200 {
		show_error("Database error " + string(body) + " for request " + endpoint)
	}

	return string(body)
}

func get_entries() EntryList {
	entries := db_query(get_entries_endpoint)
	entries = "{ \"entries\": [ " + entries[1:len(entries)-1] + " ] }"

	var entry_list EntryList
	json.Unmarshal([]byte(entries), &entry_list)

	if len(entry_list.Entries) == 0 {
		add_entry("New Entry ", "")
		return get_entries()
	}

	return entry_list
}

func add_entry(title, content string) {
	db_query(add_entry_endpoint + "?" + "title=" + url.QueryEscape(title) + "&content=" + url.QueryEscape(content))
}

func delete_entry(entry Entry) {
	db_query(delete_entry_endpoint + "?" + "id=" + url.QueryEscape(entry.Id))
}

func modify_entry(entry Entry) {
	db_query(modify_entry_endpoint + "?id=" + url.QueryEscape(entry.Id) + "&title=" + url.QueryEscape(entry.Title) + "&content=" + url.QueryEscape(entry.Content))
}

func update_notes_list(new_entries EntryList) {
	notes_list := make([]string, len(new_entries.Entries))

	for i := 0; i < len(new_entries.Entries); i++ {
		notes_list[i] = new_entries.Entries[i].Title
	}

	notes.Rows = notes_list
}

func edit_component(char string) {
	if view == Editor {
		edit_text(char, &content.Text)
	} else if view == EditTitle {
		edit_text(char, &notes.Rows[notes.SelectedRow])
	}
}

func edit_title() {
	view = EditTitle
	caret_pos = len(notes.Rows[notes.SelectedRow]) + 1
	raw_text = notes.Rows[notes.SelectedRow]
	notes.SelectedRowStyle = ui.NewStyle(ui.ColorGreen)
	edit_text("", &notes.Rows[notes.SelectedRow])
}

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	url, _ := ioutil.ReadFile("url.txt")
	base_url = string(url)
	entry_list = get_entries()

	header.Text = " [Zelara's Notebook](fg:yellow,mod:bold)"

	notes.Title = "[Notes]"
	notes.WrapText = false
	notes.SelectedRowStyle = ui.NewStyle(ui.ColorYellow)
	update_notes_list(entry_list)

	content.Title = "[Content]"
	content.Text = entry_list.Entries[0].Content

	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0/10,
			header,
		),
		ui.NewRow(9.0/10,
			ui.NewCol(1.0/5, notes),
			ui.NewCol(4.0/5, content),
		),
	)

	ui.Render(grid)

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "<C-c>":
			return
		case "<Escape>":
			view = Browser
			content.TextStyle = ui.NewStyle(ui.ColorWhite)
			notes.SelectedRowStyle = ui.NewStyle(ui.ColorYellow)
			content.Text = raw_text
			entry_list.Entries[notes.SelectedRow].Content = raw_text
			modify_entry(entry_list.Entries[notes.SelectedRow])
		case "<Enter>":
			if view == EditTitle {
				entry_list.Entries[notes.SelectedRow].Title = raw_text

				if len(strings.TrimSpace(entry_list.Entries[notes.SelectedRow].Title)) == 0 {
					entry_list.Entries[notes.SelectedRow].Title = "New Entry "
				}

				modify_entry(entry_list.Entries[notes.SelectedRow])
				update_notes_list(entry_list)
				raw_text = entry_list.Entries[notes.SelectedRow].Content
				content.Text = raw_text
				notes.SelectedRowStyle = ui.NewStyle(ui.ColorYellow)
				view = Browser
			} else if view == Browser {
				caret_pos = len(entry_list.Entries[notes.SelectedRow].Content)
				content.TextStyle = ui.NewStyle(ui.ColorYellow)
				notes.SelectedRowStyle = ui.NewStyle(ui.ColorGreen)
				raw_text = content.Text
				view = Editor
				edit_component("")
			} else {
				edit_component(e.ID)
			}
		case "<Up>":
			if view == Browser {
				notes.ScrollUp()
				content.Text = entry_list.Entries[notes.SelectedRow].Content
			} else {
				edit_component(e.ID)
			}
		case "<Down>":
			if view == Browser {
				notes.ScrollDown()
				content.Text = entry_list.Entries[notes.SelectedRow].Content
			} else {
				edit_component(e.ID)
			}
		case "<Resize>":
			payload := e.Payload.(ui.Resize)
			grid.SetRect(0, 0, payload.Width, payload.Height)
			ui.Clear()
		case "n":
			if view == Browser {
				add_entry("New Entry ", " ")
				entry_list = get_entries()
				update_notes_list(entry_list)
				content.Text = ""
				notes.SelectedRow = len(notes.Rows) - 1
				edit_title()
			} else {
				edit_component(e.ID)
			}
		case "e":
			if view == Browser {
				edit_title()
			} else {
				edit_component(e.ID)
			}
		case "d":
			if view == Browser {
				delete_entry(entry_list.Entries[notes.SelectedRow])
				entry_list = get_entries()
				update_notes_list(entry_list)

				if notes.SelectedRow == len(notes.Rows) {
					notes.SelectedRow = len(notes.Rows) - 1
				}

				content.Text = entry_list.Entries[notes.SelectedRow].Content
			} else {
				edit_component(e.ID)
			}
		default:
			edit_component(e.ID)
		}
		ui.Render(grid)
	}
}
