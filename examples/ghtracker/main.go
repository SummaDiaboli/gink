// Package main is a GitHub tracker TUI — browse your repositories and inspect
// their issues, pull requests, and recent commits without leaving the terminal.
// Requires the gh CLI to be installed and authenticated.
//
// Tab/Shift-Tab cycles focus. Up/Down navigates lists and tables.
// Left/Right switches the detail view tab or scrolls wide tables.
// Press Esc to quit.
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/SummaDiaboli/gink"
)

// ── Shared context ────────────────────────────────────────────────────────────

// SelectedRepoCtx holds the full name ("owner/repo") of the currently selected
// repository. RepoList writes to it; DetailPanel reads from it.
var SelectedRepoCtx = gink.NewContext("")

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	titleStyle   = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	sectionStyle = gink.NewStyle().Bold().Foreground(gink.ColorBrightCyan)
	hintStyle    = gink.NewStyle().Foreground(gink.ColorWhite)
	dimStyle     = gink.NewStyle().Foreground(gink.ColorWhite)
	valueStyle   = gink.NewStyle().Foreground(gink.ColorBrightWhite)
	errStyle     = gink.NewStyle().Bold().Foreground(gink.ColorBrightRed)
)

// ── Column definitions ────────────────────────────────────────────────────────

var issueCols = []gink.Column{
	{Header: "#", MaxWidth: 6},
	{Header: "Title", MinWidth: 36},
	{Header: "State", MaxWidth: 8},
	{Header: "Author", MaxWidth: 16},
	{Header: "Age", MaxWidth: 5},
}

var prCols = []gink.Column{
	{Header: "#", MaxWidth: 6},
	{Header: "Title", MinWidth: 36},
	{Header: "State", MaxWidth: 8},
	{Header: "Author", MaxWidth: 16},
	{Header: "Age", MaxWidth: 5},
	{Header: "Draft", MaxWidth: 6},
}

var commitCols = []gink.Column{
	{Header: "SHA", MaxWidth: 8},
	{Header: "Message", MinWidth: 44, MaxWidth: 72},
	{Header: "Author", MaxWidth: 18},
	{Header: "Age", MaxWidth: 5},
}

var tabOptions = []string{"Commits", "PRs", "Issues"}

// ── RepoList ──────────────────────────────────────────────────────────────────

// RepoList renders the left panel: a scrollable list of the authenticated
// user's repositories. Selecting a repo (Enter or click) updates SelectedRepoCtx.
func RepoList() gink.Element {
	focused := gink.UseFocusWithin()
	currentRepo := gink.UseContext(SelectedRepoCtx)

	repos, loading, fetchErr := gink.UseAsync(func() ([]Repo, error) {
		return fetchRepos()
	}, []any{})

	sel, setSel := gink.UseState(0)

	// Auto-select the first repository once the list arrives.
	gink.UseEffect(func() func() {
		if len(repos) > 0 && currentRepo == "" {
			gink.SetContext(SelectedRepoCtx, repos[0].NameWithOwner)
		}
		return nil
	}, []any{len(repos)})

	borderStyle := sectionStyle
	if focused {
		borderStyle = valueStyle
	}

	var body gink.Element
	switch {
	case loading:
		body = gink.Row(gink.C(gink.Spinner), gink.Text("  Loading…", dimStyle))
	case fetchErr != nil:
		body = gink.Text("error: "+fetchErr.Error(), errStyle)
	case len(repos) == 0:
		body = gink.Text("No repositories found.", dimStyle)
	default:
		items := make([]string, len(repos))
		for i, r := range repos {
			prefix := "  "
			if r.IsPrivate {
				prefix = "⊘ "
			}
			items[i] = prefix + r.Name
		}
		body = gink.C(gink.NewList(items, sel, func(i int) {
			setSel(i)
			gink.SetContext(SelectedRepoCtx, repos[i].NameWithOwner)
		}, 16))
	}

	return gink.BorderWithTitle("Repositories", gink.PaddingXY(1, 0, body), borderStyle)
}

// ── DetailPanel ───────────────────────────────────────────────────────────────

// DetailPanel renders the right panel. It fetches issues, PRs, and commits for
// the selected repo concurrently so tab switching is instant after first load.
func DetailPanel() gink.Element {
	focused := gink.UseFocusWithin()
	repo := gink.UseContext(SelectedRepoCtx)

	tab, setTab := gink.UseState("Commits")
	issueSel, setIssueSel := gink.UseState(0)
	prSel, setPRSel := gink.UseState(0)
	commitSel, setCommitSel := gink.UseState(0)

	issues, issLoading, issErr := gink.UseAsync(func() ([]Issue, error) {
		if repo == "" {
			return nil, nil
		}
		return fetchIssues(repo)
	}, []any{repo})

	prs, prLoading, prErr := gink.UseAsync(func() ([]PR, error) {
		if repo == "" {
			return nil, nil
		}
		return fetchPRs(repo)
	}, []any{repo})

	commits, cmLoading, cmErr := gink.UseAsync(func() ([]Commit, error) {
		if repo == "" {
			return nil, nil
		}
		return fetchCommits(repo)
	}, []any{repo})

	// Reset selections and tab when the active repo changes.
	gink.UseEffect(func() func() {
		setTab("Commits")
		setIssueSel(0)
		setPRSel(0)
		setCommitSel(0)
		return nil
	}, []any{repo})

	borderStyle := sectionStyle
	if focused {
		borderStyle = valueStyle
	}

	title := repo
	if title == "" {
		title = "Details"
	}

	if repo == "" {
		return gink.BorderWithTitle(title,
			gink.PaddingXY(2, 1, gink.Text("Select a repository from the list.", dimStyle)),
			borderStyle,
		)
	}

	tabRow := gink.Row(
		gink.Text("View  ", dimStyle),
		gink.C(gink.NewSelect(tabOptions, tab, setTab)),
	)

	var content gink.Element
	switch tab {
	case "Commits":
		content = renderCommits(commits, cmLoading, cmErr, commitSel, setCommitSel)
	case "PRs":
		content = renderPRs(prs, prLoading, prErr, prSel, setPRSel)
	default:
		content = renderIssues(issues, issLoading, issErr, issueSel, setIssueSel)
	}

	return gink.BorderWithTitle(title,
		gink.PaddingXY(1, 0, gink.BoxWithGap(1, tabRow, content)),
		borderStyle,
	)
}

func renderIssues(issues []Issue, loading bool, err error, sel int, setSel func(int)) gink.Element {
	switch {
	case loading:
		return gink.Row(gink.C(gink.Spinner), gink.Text("  Loading issues…", dimStyle))
	case err != nil:
		return gink.Text("error: "+err.Error(), errStyle)
	case len(issues) == 0:
		return gink.Text("No issues.", dimStyle)
	}
	rows := make([][]string, len(issues))
	for i, iss := range issues {
		rows[i] = []string{
			fmt.Sprintf("#%d", iss.Number),
			iss.Title,
			iss.State,
			iss.Author.Login,
			timeAgo(iss.CreatedAt),
		}
	}
	return gink.C(gink.NewTable(issueCols, rows, sel, setSel, 12))
}

func renderPRs(prs []PR, loading bool, err error, sel int, setSel func(int)) gink.Element {
	switch {
	case loading:
		return gink.Row(gink.C(gink.Spinner), gink.Text("  Loading pull requests…", dimStyle))
	case err != nil:
		return gink.Text("error: "+err.Error(), errStyle)
	case len(prs) == 0:
		return gink.Text("No pull requests.", dimStyle)
	}
	rows := make([][]string, len(prs))
	for i, pr := range prs {
		draft := ""
		if pr.IsDraft {
			draft = "yes"
		}
		rows[i] = []string{
			fmt.Sprintf("#%d", pr.Number),
			pr.Title,
			pr.State,
			pr.Author.Login,
			timeAgo(pr.CreatedAt),
			draft,
		}
	}
	return gink.C(gink.NewTable(prCols, rows, sel, setSel, 12))
}

func renderCommits(commits []Commit, loading bool, err error, sel int, setSel func(int)) gink.Element {
	switch {
	case loading:
		return gink.Row(gink.C(gink.Spinner), gink.Text("  Loading commits…", dimStyle))
	case err != nil:
		return gink.Text("error: "+err.Error(), errStyle)
	case len(commits) == 0:
		return gink.Text("No commits.", dimStyle)
	}
	rows := make([][]string, len(commits))
	for i, c := range commits {
		rows[i] = []string{c.SHA, c.Message, c.Author, timeAgo(c.Date)}
	}
	return gink.C(gink.NewTable(commitCols, rows, sel, setSel, 12))
}

// ── App ───────────────────────────────────────────────────────────────────────

// App is the root component. It composes the repo list and detail panel
// side by side, topped with a header and pinned footer.
func App() gink.Element {
	username, _, _ := gink.UseAsync(func() (string, error) {
		return fetchUsername()
	}, []any{})

	userLabel := username
	if userLabel == "" {
		userLabel = "…"
	}

	main := gink.PaddingXY(2, 1,
		gink.BoxWithGap(1,
			gink.Row(
				gink.Text("◆ GitHub Tracker", titleStyle),
				gink.Text("   @"+userLabel, dimStyle),
			),
			gink.C(gink.Divider),
			gink.Row(
				gink.C(RepoList),
				gink.Text("  "),
				gink.C(DetailPanel),
			),
		),
	)

	footer := gink.PaddingXY(2, 0,
		gink.Text("Tab: focus  ·  ↑↓: navigate  ·  Enter: select  ·  ←→: switch view / scroll  ·  Esc: quit", hintStyle),
	)

	return gink.AppShell(main, footer)
}

func main() {
	if err := gink.Render(App); err != nil {
		log.Fatal(err)
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// timeAgo returns a compact human-readable duration since t.
func timeAgo(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		return fmt.Sprintf("%dmo", int(d.Hours()/(24*30)))
	}
}
