<div align="center">
  <img src="public/logo/logo.gif" alt="Longitudinal Depression Trajectories CLI" width="820">
</div>

<p align="center"><big><big><strong>L</strong>ongitudinal <strong>D</strong>epression <strong>T</strong>rajectories <em>CLI</em></big></big></p>

<div align="center">
  <img src="https://img.shields.io/static/v1?label=Go&message=1.25%2B&color=00ADD8&style=for-the-badge&logo=go&logoColor=white" alt="Go 1.25+">
  <img src="https://img.shields.io/static/v1?label=Bubble%20Tea&message=TUI&color=F97316&style=for-the-badge" alt="Bubble Tea TUI">
  <img src="https://img.shields.io/static/v1?label=Command%20Palette&message=ready&color=111827&style=for-the-badge" alt="Command Palette ready">
  <img src="https://img.shields.io/static/v1?label=Homebrew&message=available&color=2E7D32&style=for-the-badge&logo=homebrew&logoColor=white" alt="Homebrew available">
  <img src="https://img.shields.io/static/v1?label=Scoop%20Bucket&message=available&color=0EA5E9&style=for-the-badge" alt="Scoop Bucket available">
</div>

<div align="center">
  <a href="https://github.com/Longitudinal-Depression-Toolkit/ldt-toolkit">Primary end-user repository: ldt-toolkit</a> -
  <a href="https://life-epi-psych.github.io">LEAP Group</a>
</div>

## <img src="public/icons/lucide/github.svg" width="32" alt="" /> About The Project

The `Longitudinal Depression Trajectories Toolkit (LDT-Toolkit)` initiative is designed for social, medical, and clinical researchers who work with repeated-measure data and need a stepping-stone path from raw cohort files to downstream modelling.

The initiative delivers two interconnected components. First, `ldt-toolkit` is the Python engine of tools and reproducible pipelines to accelerate exploration of longitudinal studies towards downstream modelling. Second, `ldt` (this repository) is a fully interactive Go CLI with a no-code terminal interface for running and orchestrating the toolkit from start to finish.

The toolset supports two broad lines of exploration. `Playground` methods help researchers iterate quickly on datasets across data preparation, data preprocessing, and machine learning phases. `Presets` provide stage-level reproducible pipelines for specific studies and are built to grow through community contributions.

This current repository is the CLI component of the project. For the full end-to-end toolkit experience, workflows, and primary documentation, use [`ldt-toolkit`](https://github.com/Longitudinal-Depression-Toolkit/ldt-toolkit).

## <img src="public/icons/lucide/terminal.svg" width="32" alt="" /> Setup & Launch

### <img src="public/icons/lucide/tally-1.svg" width="24" alt="" /> Homebrew

```bash
brew tap Longitudinal-Depression-Toolkit/homebrew-tap
brew install ldt
```

### <img src="public/icons/lucide/tally-2.svg" width="24" alt="" /> Scoop Bucket

```powershell
scoop bucket add longitudinal-depression-toolkit https://github.com/Longitudinal-Depression-Toolkit/scoop-bucket
scoop install ldt
```

### <img src="public/icons/lucide/tally-3.svg" width="24" alt="" /> From Source

```bash
git clone https://github.com/Longitudinal-Depression-Toolkit/CLI.git
cd CLI
make build
```

### <img src="public/icons/lucide/tally-4.svg" width="24" alt="" /> Launch It Anywhere

Homebrew and Scoop put `ldt` on `PATH` automatically.

If you built from source, run the installer target for your shell:

```bash
# bash
make install-bash

# fish
make install-fish
```

Now run from any directory:

```bash
ldt
```

## <img src="public/icons/lucide/command.svg" width="32" alt="" /> Novel Command Palette

We've made sure the UX of the CLI stays smooth at every moment. Use the command palette wherever you are to instantly launch any tool or preset of your choice, without walking through each menu level by hand.

In practice, launch `ldt`, open the palette with `:` or `Ctrl+P`, type a phrase like `build trajectories`, and press `Enter` to run it. While typing, `Tab` auto-completes, `Ctrl+H` opens history, `Ctrl+L` clears input, and `Esc` closes the palette.

## <img src="public/icons/lucide/goal.svg" width="32" alt="" /> What's Next?

Jump to the main repository for toolkit documentation, workflows, and end-user guidance:

- [`ldt-toolkit`](https://github.com/Longitudinal-Depression-Toolkit/ldt-toolkit)

## <img src="public/icons/lucide/fingerprint-pattern.svg" width="24" alt="" /> Security

Please do not share participant-level or restricted data in issues or pull requests.

Security policy and contact details:
- [`SECURITY.md`](SECURITY.md)

_Special thanks to [@charm.land](https://charm.land) for their amazing TUI framework!_