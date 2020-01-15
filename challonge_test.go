package challonge_test

import (
	"github.com/FlowingSPDG/go-challonge"
	"testing"
)

const (
	// Get API key from https://challonge.com/ja/settings/developer
	User string = ""
	Key  string = ""
)

func TestNew(t *testing.T) {
	client := challonge.New(User, Key)
	tournament, err := client.NewTournamentRequest("sample_tournament_1").Get()

	if err != nil {
		t.Fatalf("unable to retrieve tournament.\nERR : %v\n", err)
	}
	t.Logf("Tournament : %v\n", tournament)
	t.Logf("Tournament name: %s\n", tournament.Name)
	t.Logf("Tournament desc: %s\n", tournament.Description)

	matches := tournament.GetMatches()
	t.Logf("Matches : %v\n", matches)

	participant := tournament.Participants
	t.Logf("Participant : %v\n", participant)
}

func TestFinalize(t *testing.T) {
	client := challonge.New(User, Key)
	tournament, err := client.NewTournamentRequest("ProjectGIARS_TEST").Get()
	if err != nil {
		t.Fatalf("unable to retrieve tournament.\nERR : %v\n", err)
	}
	t.Logf("Tournament : %v\n", tournament)
	err = tournament.Finalize()
	if err != nil {
		t.Fatalf("unable to finish tournament.\nERR : %v\n", err)
	}
}
