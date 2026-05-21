#!/usr/bin/env uv run
#/// script
# requirements: ["requests", "tqdm"]
# ///

import sys
import argparse
import json
from pathlib import Path
from tqdm import tqdm
from github_utils import get_repos, get_file_content, cleanup_clones, ORG, logger

def collect_manifests(org, output_dir, public_only=False):
	"""Collect integration manifests from all repos."""
	repos = get_repos(org, public_only)
	logger.info(f"Found {len(repos)} repositories in organization '{org}'")

	output_path = Path(output_dir)
	output_path.mkdir(parents=True, exist_ok=True)

	manifests = {}
	for repo in tqdm(repos, desc="Collecting manifests"):
		repo_name = repo['name']
		manifest_content = get_file_content(org, repo_name, 'integration-manifest.json')

		if manifest_content:
			try:
				manifest = json.loads(manifest_content)
				output_file = output_path / f"{repo_name}.json"
				with open(output_file, 'w') as f:
					json.dump(manifest, f, indent=2)
				manifests[repo_name] = manifest
				logger.debug(f"{repo_name}: Manifest collected")
			except json.JSONDecodeError as e:
				logger.error(f"{repo_name}: Failed to parse manifest JSON: {e}")
				manifests[repo_name] = None
		else:
			manifests[repo_name] = None

	return manifests

def main():
	parser = argparse.ArgumentParser(
		description="Collect integration manifests from Keyfactor organization repositories"
	)
	parser.add_argument(
		"--outdir",
		default="./manifests",
		help="Output directory for collected manifests (default: ./manifests)"
	)
	parser.add_argument(
		"--public-only",
		action="store_true",
		help="Only fetch public repositories"
	)

	args = parser.parse_args()
	try:
		collect_manifests(ORG, args.outdir, args.public_only)
	finally:
		cleanup_clones()

if __name__ == "__main__":
	main()
