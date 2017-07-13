// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package engine

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/FactomProject/factomd/common/interfaces"
	"github.com/FactomProject/factomd/common/messages"
)

func LoadJournal(s interfaces.IState, journal string) {
	f, err := os.Open(journal)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	r := bufio.NewReaderSize(f, 4*1024)

	LoadJournalFromReader(s, r)
}

func LoadJournalFromString(s interfaces.IState, journalStr string) {
	f := strings.NewReader(journalStr)
	r := bufio.NewReaderSize(f, 4*1024)
	LoadJournalFromReader(s, r)
}

func LoadJournalFromReader(s interfaces.IState, r *bufio.Reader) {
	s.SetIsReplaying()
	defer s.SetIsDoneReplaying()

	fmt.Println("Replaying Journal")
	time.Sleep(time.Second * 5)
	fmt.Println("GO!")
	t := 0
	p := 0
	for {
		t++
		fmt.Println("total: ", t, " processed: ", p, "            \r")

		// line is empty if no more data
		line, err := r.ReadBytes('\n')
		if len(line) == 0 || err != nil {
			break
		}

		// Get the next word.  If not MsgHex:, then go to next line.
		adv, word, err := bufio.ScanWords(line, true)
		if string(word) != "MsgHex:" {
			continue // Go to next line.
		}
		line = line[adv:] // Remove "MsgHex:" from the line.

		// Remove spaces.
		adv, data, err := bufio.ScanWords(line, true)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Decode the hex
		binary, err := hex.DecodeString(string(data))
		if err != nil {
			fmt.Println(err)
			return
		}

		// Unmarshal the message.
		msg, err := messages.UnmarshalMessage(binary)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Process the message.
		s.InMsgQueue().Enqueue(msg)
		p++
		if s.InMsgQueue().Length() > 200 {
			for s.InMsgQueue().Length() > 50 {
				time.Sleep(time.Millisecond * 10)
			}
			time.Sleep(time.Millisecond * 100)
		}
	}

	//Waiting for state to process the message queue
	//before we disable "IsDoneReplaying"
	for s.InMsgQueue().Length() > 0 {
		time.Sleep(time.Millisecond * 100)
	}
}
