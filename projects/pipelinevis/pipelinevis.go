package pipelinevis

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type Node struct {
	ID    string
	Label string
	Type  string // "source", "transform", "destination"
	X     int
	Y     int
}

type Edge struct {
	From string
	To   string
}

type Pipeline struct {
	Nodes []Node
	Edges []Edge
}

type PageData struct {
	SVG      template.HTML
	HasGraph bool
}

func parsePipeline(input string) Pipeline {
	var pipeline Pipeline
	nodeIndex := make(map[string]bool)

	lines := strings.Split(input, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, "->")
		for i, part := range parts {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}

			id := strings.ToLower(strings.ReplaceAll(part, " ", "_"))

			if !nodeIndex[id] {
				nodeType := "transform"
				if i == 0 {
					nodeType = "source"
				}
				if i == len(parts)-1 {
					nodeType = "destination"
				}

				pipeline.Nodes = append(pipeline.Nodes, Node{
					ID:    id,
					Label: part,
					Type:  nodeType,
				})
				nodeIndex[id] = true
			}

			if i > 0 {
				fromPart := strings.TrimSpace(parts[i-1])
				fromID := strings.ToLower(strings.ReplaceAll(fromPart, " ", "_"))
				pipeline.Edges = append(pipeline.Edges, Edge{
					From: fromID,
					To:   id,
				})
			}
		}
	}

	return pipeline
}

func generateSVG(pipeline Pipeline) string {
	cols := make(map[string]int)
	for _, node := range pipeline.Nodes {
		switch node.Type {
		case "source":
			cols[node.ID] = 0
		case "transform":
			cols[node.ID] = 1
		case "destination":
			cols[node.ID] = 2
		}
	}

	colCount := make(map[int]int)
	positions := make(map[string][2]int)
	for _, node := range pipeline.Nodes {
		col := cols[node.ID]
		row := colCount[col]
		colCount[col]++
		x := 80 + col*220
		y := 80 + row*100
		positions[node.ID] = [2]int{x, y}
	}

	width := 700
	height := 80 + len(pipeline.Nodes)*100

	svg := fmt.Sprintf(`<svg viewBox="0 0 %d %d" xmlns="http://www.w3.org/2000/svg">`, width, height)

	for _, edge := range pipeline.Edges {
		from := positions[edge.From]
		to := positions[edge.To]
		svg += fmt.Sprintf(`
			<line x1="%d" y1="%d" x2="%d" y2="%d"
				stroke="#1c2340" stroke-width="2" marker-end="url(#arrow)"/>`,
			from[0]+110, from[1]+20,
			to[0], to[1]+20,
		)
	}

	for _, node := range pipeline.Nodes {
		pos := positions[node.ID]
		x, y := pos[0], pos[1]

		color := "#1c2340"
		textColor := "#5a6680"
		switch node.Type {
		case "source":
			color = "rgba(0,172,215,0.15)"
			textColor = "#00acd7"
		case "destination":
			color = "rgba(52,211,153,0.15)"
			textColor = "#34d399"
		case "transform":
			color = "rgba(251,191,36,0.1)"
			textColor = "#fbbf24"
		}

		svg += fmt.Sprintf(`
			<rect x="%d" y="%d" width="110" height="40" rx="8"
				fill="%s" stroke="%s" stroke-width="1"/>
			<text x="%d" y="%d" fill="%s"
				font-family="JetBrains Mono, monospace"
				font-size="11" text-anchor="middle">%s</text>`,
			x, y, color, textColor,
			x+55, y+25, textColor,
			node.Label,
		)
	}

	svg += `<defs>
		<marker id="arrow" markerWidth="10" markerHeight="7"
			refX="10" refY="3.5" orient="auto">
			<polygon points="0 0, 10 3.5, 0 7" fill="#1c2340"/>
		</marker>
	</defs>`

	svg += `</svg>`

	return svg
}

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmpl := template.Must(template.ParseFiles("projects/pipelinevis/index.html"))

		if r.Method == http.MethodGet {
			tmpl.Execute(w, PageData{HasGraph: false})
			return
		}

		input := r.FormValue("pipeline")
		if strings.TrimSpace(input) == "" {
			tmpl.Execute(w, PageData{HasGraph: false})
			return
		}

		pipeline := parsePipeline(input)
		svg := generateSVG(pipeline)

		tmpl.Execute(w, PageData{
			SVG:      template.HTML(svg),
			HasGraph: true,
		})
	}
}
