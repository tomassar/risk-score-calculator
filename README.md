# Risk Score Calculator 🏦

A beautiful, self-hosted risk assessment tool for lending decisions. Built with Go, Templ, and HTMX for optimal performance and user experience.

## ✨ Features

- **📊 Data Upload**: CSV file upload with validation
- **🎯 Risk Scoring**: Configurable rule-based evaluation system
- **📋 Borrower Management**: Clean, filterable borrower table
- **⚙️ Rule Engine**: Live rule editing with instant feedback
- **📈 Export Results**: Annotated CSV exports with decision recommendations
- **🔒 Self-Hosted**: Complete data privacy and transparency
- **🎨 Beautiful UI**: Apple-inspired design with modern UX

## 🎯 How It Works

### User Flow
1. **Upload CSV** → Upload borrower data (income, employment, age, etc.)
2. **Configure Rules** → Create scoring rules (e.g., unemployment = -40 points)
3. **Review Results** → View auto-calculated risk scores and decisions
4. **Export Data** → Download results with recommendations

### Scoring System
- **Base Score**: 100 points (everyone starts here)
- **Rules Applied**: Custom rules add/subtract points based on borrower data
- **Final Decision**:
  - **Score ≥ 80**: ✅ **ACCEPT** (Green badge)
  - **Score 60-79**: ⚠️ **REVIEW** (Yellow badge)  
  - **Score < 60**: ❌ **REJECT** (Red badge)

### Example Rules
- Unemployment: -40 points
- Low income (<$30k): -25 points  
- Part-time work: -15 points
- Young age (<26): -20 points

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Make (optional)

### Installation

1. Clone and setup:
```bash
git clone <repository>
cd credit-score-evaluator
go mod download
```

2. Run the application:
```bash
go run cmd/server/main.go
```

3. Open your browser to `http://localhost:8080`

### Docker Deployment

```bash
docker build -t risk-calculator .
docker run -p 8080:8080 risk-calculator
```

## 🏗️ Architecture

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── config/         # Configuration management
│   ├── handlers/       # HTTP handlers
│   ├── models/         # Data models
│   ├── services/       # Business logic
│   └── storage/        # Data persistence
├── web/
│   ├── templates/      # Templ templates
│   └── static/         # CSS, JS, assets
├── migrations/         # Database migrations
└── testdata/          # Sample CSV files
```

## ⚡ Technology Stack (GoTH)

This project follows the **GoTH stack** (Go + Templ + HTMX) philosophy with minimal JavaScript for optimal performance.

### 🔧 HTMX Responsibilities
**Data Operations (80% of app functionality):**
- ✅ **Form submissions** (rules creation, file uploads)
- ✅ **Dynamic content loading** (borrowers table, rules list)
- ✅ **Real-time filtering** (borrowers with debounced search)
- ✅ **Modal management** (rule editing modals)
- ✅ **CRUD operations** (create, update, delete rules)
- ✅ **Content swapping** (flash messages, partial updates)
- ✅ **Auto-refresh** (rules list after changes)

### 🎯 JavaScript Responsibilities
**Client-Side State Management (20% of app functionality):**
- ✅ **Theme switching** (dark/light mode with localStorage)
- ✅ **Mobile menu** (toggle, animations, keyboard navigation)
- ✅ **File drag & drop** (visual feedback, file validation)
- ✅ **UI enhancements** (loading states, smooth scrolling)
- ✅ **Keyboard shortcuts** (Escape key, Ctrl+K search focus)
- ✅ **Error handling** (HTMX error responses)

This distribution ensures **fast server-side rendering**, **minimal client-side complexity**, and **excellent SEO** while maintaining rich interactivity.

## 📈 Usage

1. **Upload Data**: Start by uploading a CSV file with borrower information
2. **Configure Rules**: Set up scoring rules based on your criteria
3. **Review Scores**: View calculated risk scores in the borrower table
4. **Export Results**: Download annotated results with recommendations

## 🔧 Configuration

Environment variables:
- `PORT`: Server port (default: 8080)
- `DB_PATH`: SQLite database path (default: ./data/app.db)
- `ENV`: Environment (development/production)

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `go test ./...`
5. Submit a pull request

## 📄 License

MIT License - see LICENSE file for details. 