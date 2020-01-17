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
	tournament, err := client.NewTournamentRequest("sample_tournament_1").Get()
	if err != nil {
		t.Fatalf("unable to retrieve tournament.\nERR : %v\n", err)
	}
	t.Logf("Tournament : %v\n", tournament)
	err = tournament.Finalize()
	if err != nil {
		t.Fatalf("unable to finish tournament.\nERR : %v\n", err)
	}
}

// Tournament should be finalized...
func TestGetFinalRank(t *testing.T) {
	client := challonge.New(User, Key)
	tournament, err := client.NewTournamentRequest("sample_tournament_1").WithParticipants().Get()
	if err != nil {
		t.Fatalf("unable to retrieve tournament.\nERR : %v\n", err)
	}
	t.Logf("Tournament : %v\n", tournament)
	for _, participant := range tournament.Participants {
		t.Logf("%s's FinalRank : %d\n", participant.Name, participant.FinalRank)
	}
}

func TestRandomize(t *testing.T) {
	client := challonge.New(User, Key)
	tournament, err := client.NewTournamentRequest("sample_tournament_1").WithParticipants().Get()
	if err != nil {
		t.Fatalf("unable to retrieve tournament.\nERR : %v\n", err)
	}
	t.Logf("Tournament : %v\n", tournament)
	err = tournament.Randomize()
	if err != nil {
		t.Fatalf("unable to randomize tournament.\nERR : %v\n", err)
	}
}

func TestGetTournaments(t *testing.T) {
	client := challonge.New(User, Key)
	tournaments, err := client.GetTournaments("all", "single elimination", "")
	if err != nil {
		t.Fatalf("unable to get tournaments.\nERR : %v\n", err)
	}
	t.Logf("Got %d items\n", len(tournaments))
	for i := 0; i < len(tournaments); i++ {
		t.Logf("Item %d : %v\n", i+1, tournaments[i])
	}
}
