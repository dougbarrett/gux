# Gux Claude Code Skills

This directory contains Claude Code skills for developing with the Gux framework.

## Available Skills

### gux-framework

Comprehensive guide for Gux development including:
- CLI commands (`gux init`, `gux gen`, `gux dev`, etc.)
- Application scaffolding
- API code generation with annotations
- Component library usage (45+ components)
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
