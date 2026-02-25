package selenium

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type graphqlResponse struct {
	Data struct {
		SessionsInfo struct {
			Sessions []struct {
				ID           string `json:"id"`
				Capabilities string `json:"capabilities"`
			} `json:"sessions"`
		} `json:"sessionsInfo"`
	} `json:"data"`
}

func KillSessionByName(ctx context.Context, hubURL, sessionName string) error {
	client := &http.Client{Timeout: 5 * time.Second}

	query := `{"query":"{ sessionsInfo { sessions { id capabilities } } }"}`
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		hubURL+"/graphql",
		bytes.NewBuffer([]byte(query)),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to query sessions: %w", err)
	}
	defer resp.Body.Close()

	var result graphqlResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	for _, s := range result.Data.SessionsInfo.Sessions {

		var caps map[string]interface{}
		if err := json.Unmarshal([]byte(s.Capabilities), &caps); err != nil {
			continue
		}

		if name, ok := caps["se:name"].(string); ok && name == sessionName {

			deleteURL := fmt.Sprintf("%s/session/%s", hubURL, s.ID)

			delReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, deleteURL, nil)
			delResp, err := client.Do(delReq)
			if err != nil {
				return fmt.Errorf("failed to delete session %s: %w", s.ID, err)
			}
			delResp.Body.Close()

			fmt.Println("Selenium session killed:", s.ID, sessionName)
			return nil
		}
	}

	fmt.Println("Session not found:", sessionName)
	return nil
}
