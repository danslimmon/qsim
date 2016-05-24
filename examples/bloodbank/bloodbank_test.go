package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"testing"
)

var Built bool

type bloodBankResult struct {
	Ticks               int
	Sims                int
	UnitsUsed           int
	UnitsTossed         int
	TransfusionsAborted int
	P90UnitAge          int
	AgeCounts           []int
}

func buildBloodBank() error {
	_, err := exec.Command("go", "build", "bloodbank.go").Output()
	if err != nil {
		return err
	}
	return nil
}

func testBloodBank(drawRate, maxOccupancy int, transfusionRate float64) (bloodBankResult, error) {
	if !Built {
		err := buildBloodBank()
		if err != nil {
			return bloodBankResult{}, err
		}
		Built = true
	}

	out, err := exec.Command("./bloodbank",
		fmt.Sprintf("%d", drawRate),
		fmt.Sprintf("%d", maxOccupancy),
		fmt.Sprintf("%f", transfusionRate),
		"test",
	).Output()
	if err != nil {
		return bloodBankResult{}, err
	}

	outStr := strings.TrimSpace(string(out))

	pieces := strings.Split(outStr, ",")
	rslt := bloodBankResult{}
	rslt.Ticks, err = strconv.Atoi(pieces[0])
	if err != nil {
		return bloodBankResult{}, err
	}
	rslt.Sims, err = strconv.Atoi(pieces[1])
	if err != nil {
		return bloodBankResult{}, err
	}
	rslt.UnitsUsed, err = strconv.Atoi(pieces[2])
	if err != nil {
		return bloodBankResult{}, err
	}
	rslt.UnitsTossed, err = strconv.Atoi(pieces[3])
	if err != nil {
		return bloodBankResult{}, err
	}
	rslt.TransfusionsAborted, err = strconv.Atoi(pieces[4])
	if err != nil {
		return bloodBankResult{}, err
	}
	rslt.P90UnitAge, err = strconv.Atoi(pieces[5])
	if err != nil {
		return bloodBankResult{}, err
	}
	for _, ageCountStr := range pieces[6:] {
		ageCount, err := strconv.Atoi(ageCountStr)
		if err != nil {
			return bloodBankResult{}, err
		}
		rslt.AgeCounts = append(rslt.AgeCounts, ageCount)
	}
	return rslt, nil
}

// Tests the limiting case where our maximum daily draw rate is very low.
//
// In this case, we expect a high number of aborted transfusions, a low number
// of wasted units, and young samples used.
func TestBloodBankLowDrawRate(t *testing.T) {
	t.Parallel()

	drawRate := 1
	maxOccupancy := 30
	transfusionRate := 30.0
	rslt, err := testBloodBank(drawRate, maxOccupancy, transfusionRate)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if rslt.TransfusionsAborted < 20*rslt.UnitsUsed {
		t.Logf("Expected way more aborted than successful transfusions; got %d aborted and %d successful",
			rslt.TransfusionsAborted, rslt.UnitsUsed)
		t.FailNow()
	}
	if rslt.UnitsTossed != 0 {
		t.Logf("Expected zero trashed units; got %d", rslt.UnitsTossed)
		t.FailNow()
	}

	olderThanMiddle := rslt.AgeCounts[len(rslt.AgeCounts)/2]
	if olderThanMiddle != 0 {
		t.Logf("Expected zero units older than the middle threshold; got %d", rslt.UnitsTossed)
		t.FailNow()
	}

	transfusionAttempts := rslt.TransfusionsAborted + rslt.UnitsUsed
	expAttempts := float64(rslt.Ticks) * transfusionRate
	if 0.9*float64(transfusionAttempts) > expAttempts || float64(transfusionAttempts) < 0.9*expAttempts/1440.0 {
		t.Logf("Expected total transfusion attempts to be in line with transfusion rate; got %d and expected ~%0.2f",
			transfusionAttempts, expAttempts)
		t.FailNow()
	}
}

// Tests the limiting case where our maximum daily draw rate is very high.
//
// In this case, we expect a low number of aborted transfusions, a high number
// of wasted units, and young samples used.
func TestBloodBankHighDrawRate(t *testing.T) {
	t.Parallel()

	drawRate := 25
	maxOccupancy := 100
	transfusionRate := 0.1
	rslt, err := testBloodBank(drawRate, maxOccupancy, transfusionRate)
	if err != nil {
		t.Log(err.Error())
		t.FailNow()
	}

	if 20*rslt.TransfusionsAborted > rslt.UnitsUsed {
		t.Logf("Expected way more successful than aborted transfusions; got %d successful and %d aborted",
			rslt.UnitsUsed, rslt.TransfusionsAborted)
		t.FailNow()
	}
	if rslt.UnitsTossed < 20*rslt.UnitsUsed {
		t.Logf("Expected a ton of trashed units; got %d tossed and %d used",
			rslt.UnitsTossed, rslt.UnitsUsed)
		t.FailNow()
	}

	olderThanMiddle := rslt.AgeCounts[len(rslt.AgeCounts)/2]
	if olderThanMiddle != 0 {
		t.Logf("Expected zero units older than the middle threshold; got %d", rslt.UnitsTossed)
		t.FailNow()
	}

	transfusionAttempts := rslt.TransfusionsAborted + rslt.UnitsUsed
	expAttempts := float64(rslt.Ticks) * transfusionRate / 1440.0
	// Tolerance is pretty loose because we don't get to pick very many random numbers in this sim.
	if 0.5*float64(transfusionAttempts) > expAttempts || float64(transfusionAttempts) < 0.5*expAttempts {
		t.Logf("Expected total transfusion attempts to be in line with transfusion rate; got %d and expected ~%0.2f",
			transfusionAttempts, expAttempts)
		t.FailNow()
	}
}
