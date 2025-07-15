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
	v := vcon.New()
	v.Subject = "Complex Call Scenario"

	// Add multiple parties
	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
		Role: "agent",
		Tel:  "tel:+12025551000",
	})
	customerIdx := v.AddParty(vcon.Party{
		Name: "John Customer",
		Role: "customer",
		Tel:  "tel:+12025552000",
	})
	supervisorIdx := v.AddParty(vcon.Party{
		Name: "Jane Supervisor",
		Role: "supervisor",
		Tel:  "tel:+12025553000",
	})
	transfereeIdx := v.AddParty(vcon.Party{
		Name: "Bob Support",
		Role: "support",
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
		Body:       "base64encodedaudiocontent...",
		Encoding:   "base64",
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
		Type:          "transfer",
		StartTime:     &threeMinLater,
		Transferee:    customerIdx,
		Transferor:    agentIdx,
		TransferTarget: transfereeIdx,
		TargetDialog:  initialCallIdx,
	})

	// A follow-up dialog with the transferee
	followupDialogIdx := v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &fourMinLater,
		Duration:   120.0, // 2 minutes
		Parties:    []int{transfereeIdx, customerIdx},
		Originator: transfereeIdx,
		MediaType:  "audio/wav",
		Body:       "base64encodedaudiocontent2...",
		Encoding:   "base64",
	})

	// Add an attachment related to the initial call
	attachmentIdx := v.AddAttachment(vcon.Attachment{
		DialogIdx: initialCallIdx,
		PartyIdx:  agentIdx,
		StartTime: now,
		MediaType: "application/pdf",
		Filename:  "customer_notes.pdf",
		Body:      "base64encodedpdfcontent...",
		Encoding:  "base64",
	})

	// Add analysis for both calls
	transcriptIdx := v.AddAnalysis(vcon.Analysis{
		Type:       "transcript",
		Dialog:     []int{initialCallIdx, followupDialogIdx},
		MediaType:  "text/plain",
		Vendor:     "TranscriptCo",
		Product:    "AutoTranscribe v1.0",
		Body:       "Customer: Hello, I'm having an issue...\nAgent: Let me transfer you...",
		Encoding:   "none",
	})

	sentimentIdx := v.AddAnalysis(vcon.Analysis{
		Type:       "sentiment",
		Dialog:     []int{initialCallIdx},
		MediaType:  "application/json",
		Vendor:     "EmotionAI",
		Product:    "SentimentAnalyzer v2.1",
		Body:       `{"overall": "neutral", "customer": "frustrated", "agent": "helpful"}`,
		Encoding:   "json",
	})

	// Validate the complex VCon
	valid, errors := v.IsValid()
	assert.True(t, valid, "Complex VCon should be valid")
	assert.Empty(t, errors, "There should be no validation errors")

	// Assert specific properties to ensure everything was created correctly
	assert.Equal(t, 4, len(v.Parties), "Should have 4 parties")
	assert.Equal(t, 3, len(v.Dialog), "Should have 3 dialogs")
	assert.Equal(t, 1, len(v.Attachments), "Should have 1 attachment")
	assert.Equal(t, 2, len(v.Analysis), "Should have 2 analysis entries")

	// Check relationships
	assert.Equal(t, initialCallIdx, v.Dialog[transferDialogIdx].TargetDialog, "Transfer should reference initial call")
	assert.Equal(t, initialCallIdx, v.Attachments[attachmentIdx].DialogIdx, "Attachment should reference initial call")
	assert.Equal(t, []int{initialCallIdx, followupDialogIdx}, v.Analysis[transcriptIdx].Dialog, "Transcript should reference both calls")
	
	// Verify sentiment analysis properties
	assert.Equal(t, "sentiment", v.Analysis[sentimentIdx].Type, "Sentiment analysis should have correct type")
	assert.Equal(t, []int{initialCallIdx}, v.Analysis[sentimentIdx].Dialog, "Sentiment analysis should reference initial call")
	assert.Equal(t, "EmotionAI", v.Analysis[sentimentIdx].Vendor, "Should have correct vendor")
	assert.Equal(t, "SentimentAnalyzer v2.1", v.Analysis[sentimentIdx].Product, "Should have correct product")
	assert.Equal(t, "json", v.Analysis[sentimentIdx].Encoding, "Should have correct encoding")
	assert.Contains(t, v.Analysis[sentimentIdx].Body, "frustrated", "Sentiment analysis should contain expected content")
}

// TestInvalidPartyReference tests validation of invalid party references
func TestInvalidPartyReference(t *testing.T) {
	// Create a VCon with a dialog referencing non-existent party
	v := vcon.New()
	v.Subject = "Invalid Party Reference Test"
	
	// Add only one party
	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
		Role: "agent",
	})
	
	// Reference a non-existent party index in dialog
	now := time.Now().UTC()
	v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Parties:    []int{agentIdx, 5}, // 5 is an invalid index
		Originator: agentIdx,
	})
	
	// Validation should fail
	valid, errors := v.IsValid()
	assert.False(t, valid, "VCon with invalid party reference should be invalid")
	assert.NotEmpty(t, errors, "Should have validation errors")
	
	// Check that the specific error is detected
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
	// Create a VCon with analysis referencing non-existent dialog
	v := vcon.New()
	v.Subject = "Invalid Dialog Reference Test"
	
	// Add a party
	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
		Role: "agent",
	})
	
	// Add a dialog
	now := time.Now().UTC()
	dialogIdx := v.AddDialog(vcon.Dialog{
		Type:       "recording",
		StartTime:  &now,
		Parties:    []int{agentIdx},
		Originator: agentIdx,
	})
	
	// Reference a non-existent dialog index in analysis
	v.AddAnalysis(vcon.Analysis{
		Type:   "transcript",
		Dialog: []int{dialogIdx, 10}, // 10 is an invalid dialog index
	})
	
	// Validation should fail
	valid, errors := v.IsValid()
	assert.False(t, valid, "VCon with invalid dialog reference should be invalid")
	assert.NotEmpty(t, errors, "Should have validation errors")
	
	// Check that the specific error is detected
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
	// Create a VCon with dialog missing required start time
	v := vcon.New()
	v.Subject = "Missing Required Fields Test"
	
	// Add a party
	agentIdx := v.AddParty(vcon.Party{
		Name: "Agent Smith",
		Role: "agent",
	})
	
	// Add a dialog without a start time
	v.AddDialog(vcon.Dialog{
		Type:       "recording", // Required
		StartTime:  nil, // Missing required field
		Parties:    []int{agentIdx},
		Originator: agentIdx,
	})
	
	// Validation should fail
	valid, errors := v.IsValid()
	assert.False(t, valid, "VCon with missing required field should be invalid")
	assert.NotEmpty(t, errors, "Should have validation errors")
	
	// Check that the specific error is detected
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
	// Create a VCon for a conference call with join/leave events
	v := vcon.New()
	v.Subject = "Complex Conference Call"
	
	// Add multiple parties
	moderatorIdx := v.AddParty(vcon.Party{
		Name: "Conference Moderator",
		Role: "moderator",
	})
	participant1Idx := v.AddParty(vcon.Party{
		Name: "Alice Participant",
		Role: "participant",
	})
	participant2Idx := v.AddParty(vcon.Party{
		Name: "Bob Participant", 
		Role: "participant",
	})
	participant3Idx := v.AddParty(vcon.Party{
		Name: "Charlie Participant",
		Role: "participant",
	})
	
	// Create timestamps
	startTime := time.Now().UTC()
	p1JoinTime := startTime.Add(30 * time.Second)
	p2JoinTime := startTime.Add(1 * time.Minute)
	p1HoldTime := startTime.Add(5 * time.Minute)
	p1UnholdTime := startTime.Add(6 * time.Minute)
	p3JoinTime := startTime.Add(7 * time.Minute)
	p2DropTime := startTime.Add(10 * time.Minute)
	endTime := startTime.Add(15 * time.Minute)
	
	// Create conference dialog
	conferenceIdx := v.AddDialog(vcon.Dialog{
		Type:       "conference",
		StartTime:  &startTime,
		Duration:   (endTime.Sub(startTime)).Seconds(),
		Parties:    []int{moderatorIdx, participant1Idx, participant2Idx, participant3Idx},
		Originator: moderatorIdx,
		MediaType:  "audio/wav",
		Body:       "base64encodedconferencecall...",
		Encoding:   "base64",
		// Add party history to track join/leave/hold events
		PartyHistory: []vcon.PartyHistory{
			{
				Party: participant1Idx,
				Event: string(vcon.PartyEventJoin),
				Time:  p1JoinTime,
			},
			{
				Party: participant2Idx,
				Event: string(vcon.PartyEventJoin),
				Time:  p2JoinTime,
			},
			{
				Party: participant1Idx,
				Event: string(vcon.PartyEventHold),
				Time:  p1HoldTime,
			},
			{
				Party: participant1Idx,
				Event: string(vcon.PartyEventUnhold),
				Time:  p1UnholdTime,
			},
			{
				Party: participant3Idx,
				Event: string(vcon.PartyEventJoin),
				Time:  p3JoinTime,
			},
			{
				Party: participant2Idx,
				Event: string(vcon.PartyEventDrop),
				Time:  p2DropTime,
			},
		},
	})
	
	// Add speaker identification analysis
	v.AddAnalysis(vcon.Analysis{
		Type:      "speaker_identification",
		Dialog:    []int{conferenceIdx},
		MediaType: "application/json",
		Vendor:    "VoiceAnalytics",
		Product:   "SpeakerID v3.2",
		Body:      `{"segments": [{"start": 0, "end": 30, "speaker": 0}, {"start": 30, "end": 45, "speaker": 1}]}`,
		Encoding:  "json",
	})
	
	// Validate the complex conference scenario
	valid, errors := v.IsValid()
	assert.True(t, valid, "Complex conference scenario should be valid")
	assert.Empty(t, errors, "There should be no validation errors")
	
	// Check the number of party history events
	assert.Equal(t, 6, len(v.Dialog[conferenceIdx].PartyHistory), "Should have 6 party history events")
	
	// Verify party join/leave sequence
	assert.Equal(t, participant1Idx, v.Dialog[conferenceIdx].PartyHistory[0].Party)
	assert.Equal(t, string(vcon.PartyEventJoin), v.Dialog[conferenceIdx].PartyHistory[0].Event)
	
	assert.Equal(t, participant2Idx, v.Dialog[conferenceIdx].PartyHistory[1].Party)
	assert.Equal(t, string(vcon.PartyEventJoin), v.Dialog[conferenceIdx].PartyHistory[1].Event)
	
	assert.Equal(t, participant2Idx, v.Dialog[conferenceIdx].PartyHistory[5].Party)
	assert.Equal(t, string(vcon.PartyEventDrop), v.Dialog[conferenceIdx].PartyHistory[5].Event)
}
