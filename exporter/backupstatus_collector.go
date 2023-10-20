// mongodb_exporter
// Copyright (C) 2017 Percona LLC
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package exporter

import (
	"context"
    "strconv"
	"strings"
	"os"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type backupStatusCollector struct {
	ctx  context.Context
	base *baseCollector

	compatibleMode bool
	topologyInfo   labelsGetter
}

// newBackupStatusCollector creates a collector for statistics on backup status.
func newBackupStatusCollector(ctx context.Context, client *mongo.Client, logger *logrus.Logger, compatible bool, topology labelsGetter) *backupStatusCollector {
	return &backupStatusCollector{
		ctx:            ctx,
		base:           newBaseCollector(client, logger),
		compatibleMode: compatible,
		topologyInfo:   topology,
	}
}

func (d *backupStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	d.base.Describe(d.ctx, ch, d.collect)
}

func (d *backupStatusCollector) Collect(ch chan<- prometheus.Metric) {
	d.base.Collect(ch)
}

func (d *backupStatusCollector) collect(ch chan<- prometheus.Metric) {
	defer measureCollectTime(ch, "mongodb", "backupstatus")()

	logger := d.base.logger

	var m bson.M
	var delblank string
    filename := "/tmp/mongodbBakStatus.txt"
    content, err := os.ReadFile(filename)
    if err != nil {
		m = bson.M{"10000"}
		ch <- prometheus.NewInvalidMetric(prometheus.NewInvalidDesc(err), err)
		return
	}

	delblank = strings.Replace(string(content), " ", "", -1)
    delnewline, _ := strconv.ParseFloat(strings.Replace(delblank,"\n", "", -1), 64)
	if delnewline == 1 {
        m = bson.M{"1"}
    } else {
		m = bson.M{"0"}
	}	

	logger.Debug("backupStatus result:")
	debugResult(logger, m)

	for _, metric := range makeMetrics("", bson.M{"backupStatus": m}, d.topologyInfo.baseLabels(), d.compatibleMode) {
		ch <- metric
	}
}