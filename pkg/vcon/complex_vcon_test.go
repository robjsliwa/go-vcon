package vcon_test

import (
	"testing"
	"time"

	"github.com/robjsliwa/go-vcon/pkg/vcon"
	"github.com/stretchr/testify/assert"
)

// TestValidComplexVCon tests creation of a valid complex VCon with multiple components
func TestValidComplexVCon(t *testing.T) {
	// Create a new VCon
	v := vcon.New("example.com")
	v.Subject = "Complex Call Scenario"

	// Add multiple parties (Role removed from core, now in CC extension)
	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
		Tel:  "tel:+12025551000",
	})
	customerIdx := v.AddParty(vcon.Party{
		Name: "John Customer",
		Tel:  "tel:+12025552000",
	})
	supervisorIdx := v.AddParty(vcon.Party{
		Name: "Jane Supervisor",
		Tel:  "tel:+12025553000",
	})
	transfereeIdx := v.AddParty(vcon.Party{
		Name: "Bob Support",
		Tel:  "tel:+12025554000",
	})

	// Create timestamps for the scenario
	now := time.Now().UTC()
	oneMinLater := now.Add(1 * time.Minute)
	twoMinLater := now.Add(2 * time.Minute)
	threeMinLater := now.Add(3 * time.Minute)
	fourMinLater := now.Add(4 * time.Minute)

	// Initial call dialog between agent and customer
	initialCallIdx := v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Duration:   180.0, // 3 minutes
		Parties:    []int{agentIdx, customerIdx},
		Originator: customerIdx,
		MediaType:  "audio/wav",
		Body:       "base64urlencodedaudiocontent",
		Encoding:   "base64url",
	})

	// Add party history to the call
	v.Dialog[initialCallIdx].PartyHistory = []vcon.PartyHistory{
		{
			Party: supervisorIdx,
			Event: string(vcon.PartyEventJoin),
			Time:  oneMinLater,
		},
		{
			Party: supervisorIdx,
			Event: string(vcon.PartyEventDrop),
			Time:  twoMinLater,
		},
	}

	// A transfer dialog
	transferDialogIdx := v.AddDialog(vcon.Dialog{
		Type:           "transfer",
		StartTime:      &threeMinLater,
		Transferee:     customerIdx,
		Transferor:     agentIdx,
		TransferTarget: vcon.NewIntValue(transfereeIdx),
		TargetDialog:   vcon.NewIntValue(initialCallIdx),
	})

	// A follow-up dialog with the transferee
	followupDialogIdx := v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &fourMinLater,
		Duration:   120.0, // 2 minutes
		Parties:    []int{transfereeIdx, customerIdx},
		Originator: transfereeIdx,
		MediaType:  "audio/wav",
		Body:       "base64urlencodedaudiocontent2",
		Encoding:   "base64url",
	})

	// Add an attachment related to the initial call
	attachmentIdx := v.AddAttachment(vcon.Attachment{
		DialogIdx: initialCallIdx,
		PartyIdx:  agentIdx,
		StartTime: now,
		MediaType: "application/pdf",
		Filename:  "customer_notes.pdf",
		Body:      "base64urlencodedpdfcontent",
		Encoding:  "base64url",
	})

	// Add analysis for both calls
	transcriptIdx := v.AddAnalysis(vcon.Analysis{
		Type:      "transcript",
		Dialog:    []int{initialCallIdx, followupDialogIdx},
		MediaType: "text/plain",
		Vendor:    "TranscriptCo",
		Product:   "AutoTranscribe v1.0",
		Body:      "Customer: Hello, I'm having an issue...\nAgent: Let me transfer you...",
		Encoding:  "none",
	})

	sentimentIdx := v.AddAnalysis(vcon.Analysis{
		Type:      "sentiment",
		Dialog:    []int{initialCallIdx},
		MediaType: "application/json",
		Vendor:    "EmotionAI",
		Product:   "SentimentAnalyzer v2.1",
		Body:      `{"overall": "neutral", "customer": "frustrated", "agent": "helpful"}`,
		Encoding:  "json",
	})

	// Validate the complex VCon
	valid, errors := v.IsValid()
	assert.True(t, valid, "Complex VCon should be valid")
	assert.Empty(t, errors, "There should be no validation errors")

	// Assert specific properties
	assert.Equal(t, 4, len(v.Parties), "Should have 4 parties")
	assert.Equal(t, 3, len(v.Dialog), "Should have 3 dialogs")
	assert.Equal(t, 1, len(v.Attachments), "Should have 1 attachment")
	assert.Equal(t, 2, len(v.Analysis), "Should have 2 analysis entries")

	// Check relationships
	assert.NotNil(t, v.Dialog[transferDialogIdx].TargetDialog, "TargetDialog should be set")
	ttVal, ok := v.Dialog[transferDialogIdx].TargetDialog.AsInt()
	assert.True(t, ok)
	assert.Equal(t, initialCallIdx, ttVal, "Transfer should reference initial call")
	assert.Equal(t, initialCallIdx, v.Attachments[attachmentIdx].DialogIdx, "Attachment should reference initial call")
	assert.Equal(t, []int{initialCallIdx, followupDialogIdx}, v.Analysis[transcriptIdx].Dialog, "Transcript should reference both calls")

	// Verify sentiment analysis properties
	assert.Equal(t, "sentiment", v.Analysis[sentimentIdx].Type)
	assert.Equal(t, []int{initialCallIdx}, v.Analysis[sentimentIdx].Dialog)
	assert.Equal(t, "EmotionAI", v.Analysis[sentimentIdx].Vendor)
	assert.Equal(t, "SentimentAnalyzer v2.1", v.Analysis[sentimentIdx].Product)
	assert.Equal(t, "json", v.Analysis[sentimentIdx].Encoding)
	assert.Contains(t, v.Analysis[sentimentIdx].Body, "frustrated")
}

// TestInvalidPartyReference tests validation of invalid party references
func TestInvalidPartyReference(t *testing.T) {
	v := vcon.New("example.com")
	v.Subject = "Invalid Party Reference Test"

	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
	})

	now := time.Now().UTC()
	v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Parties:    []int{agentIdx, 5}, // 5 is an invalid index
		Originator: agentIdx,
	})

	valid, errors := v.IsValid()
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	hasReferenceError := false
	for _, err := range errors {
		if err == "dialog at index 0 references invalid party index: 5" {
			hasReferenceError = true
			break
		}
	}
	assert.True(t, hasReferenceError, "Should detect invalid party reference")
}

// TestInvalidDialogReference tests validation of invalid dialog references
func TestInvalidDialogReference(t *testing.T) {
	v := vcon.New("example.com")
	v.Subject = "Invalid Dialog Reference Test"

	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
	})

	now := time.Now().UTC()
	dialogIdx := v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Parties:    []int{agentIdx},
		Originator: agentIdx,
	})

	v.AddAnalysis(vcon.Analysis{
		Type:   "transcript",
		Dialog: []int{dialogIdx, 10}, // 10 is an invalid dialog index
	})

	valid, errors := v.IsValid()
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	hasReferenceError := false
	for _, err := range errors {
		if err == "analysis at index 0 references invalid dialog index: 10" {
			hasReferenceError = true
			break
		}
	}
	assert.True(t, hasReferenceError, "Should detect invalid dialog reference")
}

// TestMissingRequiredFields tests validation of VCons with missing required fields
func TestMissingRequiredFields(t *testing.T) {
	v := vcon.New("example.com")
	v.Subject = "Missing Required Fields Test"

	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
	})

	v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  nil, // Missing required field
		Parties:    []int{agentIdx},
		Originator: agentIdx,
	})

	valid, errors := v.IsValid()
	assert.False(t, valid)
	assert.NotEmpty(t, errors)

	hasMissingFieldError := false
	for _, err := range errors {
		if err == "dialog at index 0 missing required field: start" {
			hasMissingFieldError = true
			break
		}
	}
	assert.True(t, hasMissingFieldError, "Should detect missing required field")
}

// TestComplexConferenceScenario tests a complex conference call scenario
func TestComplexConferenceScenario(t *testing.T) {
	v := vcon.New("example.com")
	v.Subject = "Complex Conference Call"

	moderatorIdx := v.AddParty(vcon.Party{
		Name: "Conference Moderator",
	})
	participant1Idx := v.AddParty(vcon.Party{
		Name: "Alice Participant",
	})
	participant2Idx := v.AddParty(vcon.Party{
		Name: "Bob Participant",
	})
	participant3Idx := v.AddParty(vcon.Party{
		Name: "Charlie Participant",
	})

	startTime := time.Now().UTC()
	p1JoinTime := startTime.Add(30 * time.Second)
	p2JoinTime := startTime.Add(1 * time.Minute)
	p1HoldTime := startTime.Add(5 * time.Minute)
	p1UnholdTime := startTime.Add(6 * time.Minute)
	p3JoinTime := startTime.Add(7 * time.Minute)
	p2DropTime := startTime.Add(10 * time.Minute)
	endTime := startTime.Add(15 * time.Minute)

	conferenceIdx := v.AddDialog(vcon.Dialog{
		Type:       "conference",
		StartTime:  &startTime,
		Duration:   (endTime.Sub(startTime)).Seconds(),
		Parties:    []int{moderatorIdx, participant1Idx, participant2Idx, participant3Idx},
		Originator: moderatorIdx,
		MediaType:  "audio/wav",
		Body:       "base64urlencodedconferencecall",
		Encoding:   "base64url",
		PartyHistory: []vcon.PartyHistory{
			{Party: participant1Idx, Event: string(vcon.PartyEventJoin), Time: p1JoinTime},
			{Party: participant2Idx, Event: string(vcon.PartyEventJoin), Time: p2JoinTime},
			{Party: participant1Idx, Event: string(vcon.PartyEventHold), Time: p1HoldTime},
			{Party: participant1Idx, Event: string(vcon.PartyEventUnhold), Time: p1UnholdTime},
			{Party: participant3Idx, Event: string(vcon.PartyEventJoin), Time: p3JoinTime},
			{Party: participant2Idx, Event: string(vcon.PartyEventDrop), Time: p2DropTime},
		},
	})

	v.AddAnalysis(vcon.Analysis{
		Type:      "speaker_identification",
		Dialog:    []int{conferenceIdx},
		MediaType: "application/json",
		Vendor:    "VoiceAnalytics",
		Product:   "SpeakerID v3.2",
		Body:      `{"segments": [{"start": 0, "end": 30, "speaker": 0}, {"start": 30, "end": 45, "speaker": 1}]}`,
		Encoding:  "json",
	})

	valid, errors := v.IsValid()
	assert.True(t, valid)
	assert.Empty(t, errors)

	assert.Equal(t, 6, len(v.Dialog[conferenceIdx].PartyHistory))

	assert.Equal(t, participant1Idx, v.Dialog[conferenceIdx].PartyHistory[0].Party)
	assert.Equal(t, string(vcon.PartyEventJoin), v.Dialog[conferenceIdx].PartyHistory[0].Event)

	assert.Equal(t, participant2Idx, v.Dialog[conferenceIdx].PartyHistory[1].Party)
	assert.Equal(t, string(vcon.PartyEventJoin), v.Dialog[conferenceIdx].PartyHistory[1].Event)

	assert.Equal(t, participant2Idx, v.Dialog[conferenceIdx].PartyHistory[5].Party)
	assert.Equal(t, string(vcon.PartyEventDrop), v.Dialog[conferenceIdx].PartyHistory[5].Event)
}
