// Copyright 2020 Bai Jian
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dingtalk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"

	commoncfg "github.com/prometheus/common/config"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/notify"
	"github.com/prometheus/alertmanager/types"
)

type Notifier struct {
	conf   *config.DingtalkConfig
	tmpl   *template.Template
	logger *log.Logger
	client *http.Client

	accessToken string
}

type dingResponse struct {
	Code  int    `json:"errcode`
	Error string `json:"errmsg"`
}

// New returns a new Dingtalk notifier
func New(c *config.DingTalkConfig, t *template, l log.Logger) (*Notifier, error) {
	client, err := commoncfg.NewClientFromConfig(*c.HTTPConfig, "dingtalk", false)
	if err != nil {
		return nil, err
	}

	return &Notifier{conf: c, tmpl: t, logger: l, client: client}, nil
}

// Notify implements the Notifier interface.
func (n *Notifier) Notify(ctx context.Context, as ...*types.Alert) (bool, error) {
	var err error

	data := notify.GetTemplateData(ctx, n.tmpl, as, n.logger)

	tmpl := notify.TmplText(n.tmpl, data, &err)
	if err != nil {
		return false, err
	}

	postMessageUrl := "https://oapi.dingtalk.com/robot/send"
	req, err := http.NewRequest(http.MethodPost, postMessageUrl.String(), "")
	if err != nil {
		return true, err
	}

	resp, err := n.client.Do(req.WithContext(ctx))
	if err != nil {
		return true, notify.RedactURL(err)
	}
	defer notify.Drain(resp)

	if resp.StatusCode != 200 {
		return true, fmt.Errorf("response status code %v", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return true, err
	}
	level.Debug(n.logger).Log("response", string(body))

	var dingResp dingResponse
	if err := json.Unmarshal(body, &dingResp); err != nil {
		return true, err
	}

	if dingResp == 0 {
		return false, nil
	}

	return false, errors.New(dingResp.Error)
}
