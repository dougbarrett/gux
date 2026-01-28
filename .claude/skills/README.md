# Gux Claude Code Skills

This directory contains Claude Code skills for developing with the Gux framework.

## Available Skills

### gux-framework

Comprehensive guide for Gux development including:

**Core Framework (`core/` package)**:
- Universal rendering system (SSR + WASM hydration)
- Node-based component architecture
- Element helpers (Div, Button, Input, etc.)
- Reactive state management
- Hybrid routing with bundle support
- CRUD API generation with DTOs and hooks

**Components Library**:
- 45+ production-ready UI components
- Tailwind CSS integration

**Development**:
- CLI commands (`gux init`, `gux gen`, `gux dev`, `gux model`, etc.)
- Model scaffolding with auth preset (automatic password hashing)
- Customizable user roles via config
- API code generation with annotations
- State management patterns
- Server utilities
- Build and deployment

## Usage

These skills are automatically loaded by Claude Code when working in this repository.

### In Claude Code

The skill content is automatically available to Claude when you're working on Gux projects. You can also explicitly reference it:

```
Use the gux-framework skill to help me create a new component
```

### For Your Own Projects

To use these skills in projects built with Gux:

1. Copy the `.claude/skills/` directory to your project
2. Or add as a git submodule:
   ```bash
   git submodule add https://github.com/dougbarrett/gux.git .gux
   ln -s .gux/.claude/skills .claude/skills
   ```

## Contributing

To improve these skills:
1. Edit the markdown files in this directory
2. Test with Claude Code in a Gux project
3. Submit a PR with your improvements
