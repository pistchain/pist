// Copyright 2018 The PistChain Authors
// This file is part of the pist library.
//
// The pist library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The pist library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the pist library. If not, see <http://www.gnu.org/licenses/>.

package dashboard

import (
	"time"
)

// collectCommitteeData gathers data about the committee and sends it to the clients.
func (db *Dashboard) collectCommitteeData() {
	defer db.wg.Done()
	agent := db.pist.PbftAgent()

	for {
		select {
		case errc := <-db.quit:
			errc <- nil
			return
		case <-time.After(db.config.Refresh):
			number := agent.CommitteeNumber()
			isCommittee := agent.IsCommitteeMember()
			isLeader := agent.IsLeader()
			currentCommittee := agent.GetCurrentCommittee()
			backCommittee := agent.GetAlternativeCommittee()
			/*committeeNumber := &ChartEntry{
				Value: float64(number),
			}*/

			db.sendToAll(&Message{
				Committee: &CommitteeMessage{
					Number:            number,
					IsCommitteeMember: isCommittee,
					IsLeader:          isLeader,
					Committee:         currentCommittee,
					BackCommittee:     backCommittee,
				},
			})

		}
	}
}
