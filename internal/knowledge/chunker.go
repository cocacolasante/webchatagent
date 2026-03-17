package knowledge

import "strings"

const defaultChunkSize = 1000
const defaultOverlap = 100

// ChunkDocument splits a document into overlapping text chunks.
// This is a simple implementation; future versions can use semantic chunking.
func ChunkDocument(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}
	if overlap < 0 {
		overlap = defaultOverlap
	}

	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}

	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}
		chunks = append(chunks, text[start:end])
		if end == len(text) {
			break
		}
		start = end - overlap
	}
	return chunks
}
