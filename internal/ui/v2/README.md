# Refactored UI Architecture

## Philosophy

This refactored structure separates concerns clearly:

1. **Models** (`models/`) - Pure bubbletea models, only handle UI state and rendering
2. **Store** (`store/`) - Data fetching, caching, and state management
3. **Theme** (`theme/`) - All styling, colors, keybindings in one place
4. **Messages** (`messages/`) - All tea.Msg types for communication

## Key Changes

### 1. Models are Pure UI

Models should NOT:

- Make API calls directly
- Know about kubectl/k8s
- Manage complex data fetching logic

Models SHOULD:

- Handle key presses and update their own visual state
- Render themselves
- Send commands that result in messages

### 2. Data Flow

```
User Action → Model.Update() → tea.Cmd → Store.Fetch() → tea.Msg → Model.Update() → View()
```

### 3. Content Updates

Table content should be updated when:

- Initial load (via Init())
- Periodic refresh (via tick messages)
- User triggers refresh (key binding)
- Parent tells it to (via custom message)

### 4. Navigation

Instead of a stack, use an enum-based "screen" approach:

```go
type Screen int
const (
    ScreenTable Screen = iota
    ScreenViewport
    ScreenHelp
)
```

The Root model holds `currentScreen` and routes messages accordingly.

## File Structure

```
internal/ui/v2/
├── messages/
│   └── messages.go      # All message types
├── models/
│   ├── table.go         # Pure table UI model
│   ├── viewport.go      # Pure viewport UI model
│   └── root.go          # Root model, handles routing
├── store/
│   └── store.go         # Data fetching and caching
└── theme/
    ├── keys.go          # Key bindings
    └── styles.go        # All styles and colors
```
