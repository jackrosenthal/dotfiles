#!/usr/bin/env python3

"""Repo upload wrapper for CrOS."""

import argparse
import dataclasses
import logging
from pathlib import Path
import subprocess
import sys


def find_checkout() -> Path:
    for path in Path.cwd().resolve().parents:
        if (path / ".repo").is_dir():
            return path
    raise OSError("Unable to find base of checkout")


NO_VERIFIED_REPOS = {
    "chromite/infra/proto",
    "infra/proto",
    "infra/recipes",
    "src/config",
}


@dataclasses.dataclass
class GWSQ:
    address: str
    paths: list[str] = dataclasses.field(default_factory=list)


GWSQS = {
    "build": GWSQ(
        "chromeos-build-team@google.com",
        paths=[
            "chromite/**",
            "src/scripts/**",
        ],
    ),
    "ci": GWSQ(
        "chromeos-continuous-integration-team@google.com",
        paths=[
            "infra/config/**",
            "infra/recipes/**",
        ],
    ),
    "ec": GWSQ(
        "cros-ec-reviewers@google.com",
        paths=[
            "src/platform/ec/**",
        ],
    ),
    "ebuild": GWSQ(
        "ebuild-reviews@google.com",
        paths=[
            "src/overlays/**",
            "src/private-overlays/**",
            "src/third_party/chromiumos-overlay/**",
            "src/third_party/eclass-overlay/**",
            "src/third_party/portage-stable/**",
            "src/third_party/toolchain-overlay/**",
        ],
    ),
    "signing": GWSQ(
        "croskeymanagers@google.com",
        paths=["src/platform/signing/**"],
    ),
}


def find_best_gwsq() -> GWSQ:
    rel_path = Path.cwd().resolve().relative_to(find_checkout())
    garbage_file_path = rel_path / "garbage"

    candidate_gwsqs = {
        "build": 0,
    }
    for name, gwsq in GWSQS.items():
        for pattern in gwsq.paths:
            if garbage_file_path.match(pattern):
                current = candidate_gwsqs.get(name, 0)
                if len(pattern) > current:
                    candidate_gwsqs[name] = len(pattern)
    return GWSQS[max(candidate_gwsqs, key=lambda k: candidate_gwsqs[k])]


def get_project():
    for path in (Path.cwd(), *Path.cwd().parents):
        if (path / ".git").exists():
            return str(path.relative_to(find_checkout()))
    raise OSError("Not in a project")


ALIASES = {
    **{f"gwsq:{k}": v.address for k, v in GWSQS.items()},
    "engeg": "engeg@google.com",
    "navil": "navil@google.com",
}


def run_preupload():
    subprocess.run(
        [find_checkout() / "src" / "repohooks" / "pre-upload.py"],
        check=True,
    )


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--check-only",
        action="store_true",
        help="Only run pre-upload checks then exit.",
    )
    parser.add_argument(
        "--gwsq",
        action="store_true",
        help="Find an automatic GWSQ to review.",
    )
    parser.add_argument(
        "-r",
        "--re",
        action="append",
        dest="reviewers",
        default=[],
        help="Reviewer",
    )
    parser.add_argument(
        "--cc",
        action="append",
        dest="ccs",
        default=[],
        help="CC",
    )
    parser.add_argument(
        "--not-verified",
        action="store_true",
        help="Don't set Verified+1.",
    )
    parser.add_argument(
        "-l",
        "--label",
        dest="labels",
        action="append",
        default=[],
        help="Upload with label",
    )
    parser.add_argument(
        "-a",
        "--auto-submit",
        dest="labels",
        action="append_const",
        const="Auto-Submit+1",
        help="Set Auto-Submit+1",
    )
    parser.add_argument(
        "-d",
        "--cq-dry-run",
        dest="labels",
        action="append_const",
        const="Commit-Queue+1",
        help="Set Commit-Queue+1",
    )
    parser.add_argument(
        "--cq-submit",
        dest="labels",
        action="append_const",
        const="Commit-Queue+2",
        help="Set Commit-Queue+2",
    )

    opts = parser.parse_args()

    try:
        run_preupload()
    except subprocess.CalledProcessError:
        logging.error("Pre-upload checks failed.")
        if opts.check_only:
            sys.exit(1)
        response = input("Continue anyway [y/N]? ")
        if not response or response.lower() != "y":
            sys.exit(1)

    if opts.check_only:
        return

    reviewers = []

    if opts.gwsq:
        reviewers.append(find_best_gwsq().address)

    for reviewer in opts.reviewers:
        reviewers.append(ALIASES.get(reviewer, reviewer))

    ccs = [ALIASES.get(x, x) for x in opts.ccs]

    labels = list(opts.labels)
    repo = get_project()
    if (
        not opts.not_verified
        and not any(x.startswith("Verified") for x in labels)
        and repo not in NO_VERIFIED_REPOS
    ):
        labels.append("Verified+1")

    upstream = subprocess.run(
        ["git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "@{u}"],
        check=True,
        stdout=subprocess.PIPE,
        encoding="utf-8",
    ).stdout.strip()
    upstream_remote, upstream_branch = upstream.split("/")

    push_options = []
    for reviewer in reviewers:
        push_options.append(f"r={reviewer}")
    for cc in ccs:
        push_options.append(f"cc={cc}")
    for label in labels:
        push_options.append(f"l={label}")

    subprocess.run(
        [
            "git",
            "push",
            upstream_remote,
            f"HEAD:refs/for/{upstream_branch}%{','.join(push_options)}",
        ],
        check=True,
    )


if __name__ == "__main__":
    main()
