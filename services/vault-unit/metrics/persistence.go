// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metrics

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"github.com/jancajthaml-openbank/vault-unit/utils"
)

// MarshalJSON serialises Metrics as json preserving uint64
func (entity *Metrics) MarshalJSON() ([]byte, error) {
	return []byte("{\"snapshotCronLatency\":" + strconv.FormatFloat(entity.snapshotCronLatency.Percentile(0.95), 'f', -1, 64) + ",\"updatedSnapshots\":" + strconv.FormatInt(entity.updatedSnapshots.Count(), 10) + ",\"createdAccounts\":" + strconv.FormatInt(entity.createdAccounts.Count(), 10) + ",\"promisesAccepted\":" + strconv.FormatInt(entity.promisesAccepted.Count(), 10) + ",\"commitsAccepted\":" + strconv.FormatInt(entity.commitsAccepted.Count(), 10) + ",\"rollbacksAccepted\":" + strconv.FormatInt(entity.rollbacksAccepted.Count(), 10) + "}"), nil
}

// UnmarshalJSON unmarshal json of Metrics entity
func (entity *Metrics) UnmarshalJSON(data []byte) error {
	if entity == nil {
		return fmt.Errorf("cannot unmarshall to nil pointer")
	}
	all := struct {
		SnapshotCronLatency float64 `json:"snapshotCronLatency"`
		UpdatedSnapshots    int64   `json:"updatedSnapshots"`
		CreatedAccounts     int64   `json:"createdAccounts"`
		PromisesAccepted    int64   `json:"promisesAccepted"`
		CommitsAccepted     int64   `json:"commitsAccepted"`
		RollbacksAccepted   int64   `json:"rollbacksAccepted"`
	}{}
	err := utils.JSON.Unmarshal(data, &all)
	if err != nil {
		return err
	}

	entity.promisesAccepted.Clear()
	entity.promisesAccepted.Inc(all.PromisesAccepted)

	entity.commitsAccepted.Clear()
	entity.commitsAccepted.Inc(all.CommitsAccepted)

	entity.rollbacksAccepted.Clear()
	entity.rollbacksAccepted.Inc(all.RollbacksAccepted)

	entity.createdAccounts.Clear()
	entity.createdAccounts.Inc(all.CreatedAccounts)

	entity.updatedSnapshots.Mark(all.UpdatedSnapshots)

	entity.snapshotCronLatency.Update(time.Duration(all.SnapshotCronLatency))

	return nil
}

func (metrics *Metrics) Persist() error {
	if metrics == nil {
		return fmt.Errorf("cannot persist nil reference")
	}
	tempFile := metrics.output + "_temp"
	data, err := utils.JSON.Marshal(metrics)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return err
	}
	if err := os.Rename(tempFile, metrics.output); err != nil {
		return err
	}
	return nil
}

func (metrics *Metrics) Hydrate() error {
	if metrics == nil {
		return fmt.Errorf("cannot hydrate nil reference")
	}
	f, err := os.OpenFile(metrics.output, os.O_RDONLY, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	buf := make([]byte, fi.Size())
	_, err = f.Read(buf)
	if err != nil && err != io.EOF {
		return err
	}
	err = utils.JSON.Unmarshal(buf, metrics)
	if err != nil {
		return err
	}
	return nil
}
