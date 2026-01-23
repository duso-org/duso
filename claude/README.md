# Claude Skills

This directory contains Claude Code skills for the Duso project.

## Using the Duso Skill

The `skills/duso/SKILL.md` file defines a Claude Code skill that integrates Duso scripting language documentation and functionality into Claude Code.

### Installation

To use this skill with Claude Code, symlink it to your Claude skills directory:

```bash
ln -s /path/to/duso/claude/skills/duso/SKILL.md ~/.claude/skills/duso.md
```

Replace `/path/to/duso` with the full path to your Duso repository.

Alternatively, if you're in the repository root:

```bash
ln -s "$(pwd)/claude/skills/duso/SKILL.md" ~/.claude/skills/duso.md
```

Once symlinked, the skill will be available in Claude Code and can be invoked with the `/duso` command.

### About SKILL.md

The `SKILL.md` file uses frontmatter (YAML at the top between `---` markers) to define:
- `name` - The display name of the skill
- `description` - What the skill does

The rest of the file contains the skill's documentation and content, which is made available to Claude Code when the skill is invoked.
