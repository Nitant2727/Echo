package analytics

import (
	"html/template"
	"net/http"
	"sort"

	"github.com/windowmonitor/pkg/storage"
)

type Visualizer struct {
	storage *storage.Storage
}

func NewVisualizer(storage *storage.Storage) *Visualizer {
	return &Visualizer{storage: storage}
}

func (v *Visualizer) StartServer(addr string) error {
	http.HandleFunc("/", v.handleDashboard)
	http.HandleFunc("/data", v.handleData)
	return http.ListenAndServe(addr, nil)
}

type ViewData struct {
	Stats []StatData
}

type StatData struct {
	Title      string
	Minutes    float64
	Percentage float64
}

const dashboardTemplate = `
<!DOCTYPE html>
<html>
<head>
    <title>Window Usage Analytics</title>
    <style>
        :root {
            --bg-primary: #1e1e1e;
            --bg-secondary: #252526;
            --text-primary: #ffffff;
            --text-secondary: #cccccc;
            --accent-color: #0078d4;
            --border-color: #404040;
            --hover-color: #2a2d2e;
        }
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, sans-serif;
            background-color: var(--bg-primary);
            color: var(--text-primary);
            line-height: 1.5;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 20px;
        }
        .header {
            padding: 20px 0;
            border-bottom: 1px solid var(--border-color);
            margin-bottom: 24px;
        }
        .header h1 {
            font-size: 24px;
            font-weight: 500;
        }
        .chart {
            background-color: var(--bg-secondary);
            border-radius: 8px;
            padding: 24px;
            box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }
        .chart-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
        }
        .chart-title {
            font-size: 18px;
            font-weight: 500;
            color: var(--text-primary);
        }
        .stats-grid {
            display: grid;
            gap: 12px;
        }
        .stat-item {
            background-color: var(--bg-primary);
            border: 1px solid var(--border-color);
            border-radius: 6px;
            padding: 16px;
            transition: background-color 0.2s;
        }
        .stat-item:hover {
            background-color: var(--hover-color);
        }
        .stat-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        .stat-title {
            font-weight: 500;
            color: var(--text-primary);
            white-space: nowrap;
            overflow: hidden;
            text-overflow: ellipsis;
            max-width: 70%;
        }
        .stat-time {
            color: var(--text-secondary);
            font-size: 14px;
        }
        .progress-bar {
            height: 6px;
            background-color: var(--border-color);
            border-radius: 3px;
            overflow: hidden;
        }
        .progress-fill {
            height: 100%;
            background-color: var(--accent-color);
            border-radius: 3px;
            transition: width 0.3s ease;
        }
        .stat-details {
            display: flex;
            justify-content: space-between;
            margin-top: 8px;
            font-size: 14px;
            color: var(--text-secondary);
        }
        @media (max-width: 768px) {
            .container {
                padding: 16px;
            }
            .chart {
                padding: 16px;
            }
            .stat-header {
                flex-direction: column;
                align-items: flex-start;
            }
            .stat-title {
                max-width: 100%;
                margin-bottom: 4px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Window Usage Analytics</h1>
        </div>
        <div class="chart">
            <div class="chart-header">
                <h2 class="chart-title">Most Active Windows (Last 24 Hours)</h2>
            </div>
            <div class="stats-grid">
                {{range .Stats}}
                <div class="stat-item">
                    <div class="stat-header">
                        <div class="stat-title">{{.Title}}</div>
                        <div class="stat-time">{{printf "%.1f" .Minutes}} min</div>
                    </div>
                    <div class="progress-bar">
                        <div class="progress-fill" style="width: {{.Percentage}}%;"></div>
                    </div>
                    <div class="stat-details">
                        <span>Usage</span>
                        <span>{{printf "%.1f" .Percentage}}%</span>
                    </div>
                </div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>
`

func (v *Visualizer) handleDashboard(w http.ResponseWriter, r *http.Request) {
	stats, err := v.storage.GetDailyStats()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Sort stats by duration
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Duration > stats[j].Duration
	})

	// Take top 10 most used windows
	if len(stats) > 10 {
		stats = stats[:10]
	}

	// Calculate total duration and prepare view data
	var totalDuration float64
	for _, stat := range stats {
		totalDuration += stat.Duration.Minutes()
	}

	viewData := ViewData{
		Stats: make([]StatData, len(stats)),
	}

	for i, stat := range stats {
		minutes := stat.Duration.Minutes()
		percentage := (minutes / totalDuration) * 100
		viewData.Stats[i] = StatData{
			Title:      stat.Title,
			Minutes:    minutes,
			Percentage: percentage,
		}
	}

	// Parse and execute template
	tmpl, err := template.New("dashboard").Parse(dashboardTemplate)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, viewData); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (v *Visualizer) handleData(w http.ResponseWriter, r *http.Request) {
	// Redirect to dashboard since we're not using AJAX anymore
	http.Redirect(w, r, "/", http.StatusFound)
}
