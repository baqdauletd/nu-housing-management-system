package analysis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

const defaultOpenAIModel = "gpt-5.4"

type OpenAIAnalyzer struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewOpenAIAnalyzer(apiKey, model string) *OpenAIAnalyzer {
	if strings.TrimSpace(model) == "" {
		model = defaultOpenAIModel
	}

	return &OpenAIAnalyzer{
		apiKey:  strings.TrimSpace(apiKey),
		model:   model,
		baseURL: "https://api.openai.com/v1/responses",
		httpClient: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (a *OpenAIAnalyzer) AnalyzeDocument(ctx context.Context, req Request) (Result, error) {
	extractedText, err := ExtractTextFromPDF(req.PDFBytes)
	if err != nil {
		extractedText = ""
	}
	if a.apiKey == "" {
		return ManualReviewResult(req, BuildPreview(extractedText), "OpenAI analysis is not configured, so this document requires manual review.", "openai_not_configured"), fmt.Errorf("openai api key not configured")
	}

	rawJSON, parsed, err := a.callOpenAI(ctx, req, extractedText)
	if err != nil {
		summary := "Automated analysis failed, so this document requires manual review."
		if strings.TrimSpace(extractedText) == "" {
			summary = "The uploaded PDF could not be processed automatically, including OCR for scanned pages, so it requires manual review."
		}
		summary = summary + " OpenAI error: " + strings.TrimSpace(err.Error())
		return ManualReviewResult(req, BuildPreview(extractedText), summary, "analysis_request_failed"), err
	}

	return PostProcessResult(req, extractedText, parsed, rawJSON), nil
}

func (a *OpenAIAnalyzer) callOpenAI(ctx context.Context, req Request, extractedText string) (string, aiResponse, error) {
	fileID, err := a.uploadPDFFile(ctx, req.PDFBytes)
	if err != nil {
		return "", aiResponse{}, err
	}

	requestBody := map[string]any{
		"model": a.model,
		"input": []map[string]any{
			{
				"role": "system",
				"content": []map[string]string{
					{"type": "input_text", "text": analysisSystemPrompt()},
				},
			},
			{
				"role": "user",
				"content": []map[string]any{
					{
						"type":    "input_file",
						"file_id": fileID,
					},
					{
						"type": "input_text",
						"text": MustJSON(analysisPayload(req, extractedText)),
					},
				},
			},
		},
		"text": map[string]any{
			"format": map[string]any{
				"type":   "json_schema",
				"name":   "document_analysis",
				"strict": true,
				"schema": analysisJSONSchema(),
			},
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return "", aiResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", aiResponse{}, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", aiResponse{}, fmt.Errorf("openai request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", aiResponse{}, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", aiResponse{}, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBytes)))
	}

	outputText, err := extractResponseOutputText(responseBytes)
	if err != nil {
		return "", aiResponse{}, err
	}

	var parsed aiResponse
	if err := json.Unmarshal([]byte(outputText), &parsed); err != nil {
		return outputText, aiResponse{}, fmt.Errorf("decode structured output: %w", err)
	}
	return outputText, parsed, nil
}

func (a *OpenAIAnalyzer) uploadPDFFile(ctx context.Context, pdfBytes []byte) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("purpose", "user_data"); err != nil {
		return "", err
	}

	fileWriter, err := writer.CreateFormFile("file", "document.pdf")
	if err != nil {
		return "", err
	}
	if _, err := fileWriter.Write(pdfBytes); err != nil {
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.openai.com/v1/files", &body)
	if err != nil {
		return "", err
	}
	httpReq.Header.Set("Authorization", "Bearer "+a.apiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := a.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("openai file upload failed: %w", err)
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return "", fmt.Errorf("openai file upload returned status %d: %s", resp.StatusCode, strings.TrimSpace(string(responseBytes)))
	}

	var fileResp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(responseBytes, &fileResp); err != nil {
		return "", err
	}
	if strings.TrimSpace(fileResp.ID) == "" {
		return "", fmt.Errorf("openai file upload response missing file id")
	}
	return fileResp.ID, nil
}

func extractResponseOutputText(responseBytes []byte) (string, error) {
	var raw map[string]any
	if err := json.Unmarshal(responseBytes, &raw); err != nil {
		return "", err
	}
	if outputText, ok := raw["output_text"].(string); ok && strings.TrimSpace(outputText) != "" {
		return outputText, nil
	}

	output, ok := raw["output"].([]any)
	if !ok {
		return "", fmt.Errorf("responses output missing")
	}

	for _, item := range output {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		content, ok := itemMap["content"].([]any)
		if !ok {
			continue
		}
		for _, entry := range content {
			entryMap, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := entryMap["text"].(string); ok && strings.TrimSpace(text) != "" {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("responses output text missing")
}
