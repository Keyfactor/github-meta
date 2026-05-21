#!/usr/bin/env uv run
#/// script
# requirements: ["requests", "tqdm"]
# ///

import sys
import argparse
import json
import csv
import base64
import requests
from pathlib import Path
from tqdm import tqdm
from github_utils import (
	get_headers, get_repos, get_file_content, is_public_repo,
	get_csproj_versions, get_workflow_actions_version, get_readme_last_commit,
	cleanup_clones, ORG, GITHUB_API, logger
)


def load_manifests_from_directory(manifest_dir):
	"""Load all manifest files from a directory."""
	manifests = {}
	manifest_path = Path(manifest_dir)

	if not manifest_path.exists():
		return manifests

	for file in manifest_path.glob("*.json"):
		repo_name = file.stem
		try:
			with open(file) as f:
				manifests[repo_name] = json.load(f)
		except json.JSONDecodeError:
			manifests[repo_name] = None

	return manifests

def load_manifests_from_github(public_only=False):
	"""Fetch manifests directly from GitHub for all repos."""
	repos = get_repos(ORG, public_only)
	logger.info(f"Found {len(repos)} repositories in organization '{ORG}'")

	manifests = {}
	for repo in tqdm(repos, desc="Fetching manifests from GitHub"):
		repo_name = repo['name']
		url = f"{GITHUB_API}/repos/{ORG}/{repo_name}/contents/integration-manifest.json"
		resp = requests.get(url, headers=get_headers())

		if resp.status_code == 200:
			content = resp.json().get('content')
			if content:
				try:
					manifest_content = base64.b64decode(content).decode('utf-8')
					manifest = json.loads(manifest_content)
					manifests[repo_name] = manifest
				except (json.JSONDecodeError, Exception):
					manifests[repo_name] = None

	return manifests

def build_inventory(manifest_dir=None, public_only=False, integration_type=None):
	"""Build inventory from manifests."""
	if manifest_dir:
		manifests = load_manifests_from_directory(manifest_dir)
	else:
		manifests = load_manifests_from_github(public_only)

	inventory = []

	for repo_name, manifest in tqdm(manifests.items(), desc="Processing manifests"):
		if manifest is None:
			continue

		# Apply integration_type filter
		if integration_type and manifest.get('integration_type') != integration_type:
			continue

		# Check if repo is public (if filtering)
		if public_only and not is_public_repo(ORG, repo_name):
			continue

		# Extract .NET versions from release_project
		# net_versions = ""
		release_project = manifest.get('release_project', None)
		net_versions = ",".join(get_csproj_versions(ORG, repo_name, release_project))

		# Build inventory row
		row = {
			"Integration Name": manifest.get('name', '[UNKNOWN]'),
			"Integration Type": manifest.get('integration_type', '[UNKNOWN]'),
			"Feature Flags": "",
			".NET Versions": net_versions,
			"Store Types": "",
			"keyfactor/actions Version": get_workflow_actions_version(ORG, repo_name) or "",
			"GitHub Repo URL": f"https://github.com/{ORG}/{repo_name}",
			"Confluence Link": "",
			"Last README.md Update": "",
		}

		# Handle store types for orchestrator integrations
		about = manifest.get('about', {})
		if 'orchestrator' in about:
			store_types = about['orchestrator'].get('store_types', [])
			try:
				if store_types and isinstance(store_types, list) and all(isinstance(st, dict) for st in store_types):
					store_types = [st['ShortName'] for st in store_types if isinstance(st, dict) and 'ShortName' in st]
				elif store_types and isinstance(store_types, dict):
					store_types = [t['ShortName'] for t in store_types.values() if isinstance(t, dict) and 'ShortName' in t]
				
				store_types = [st for st in store_types]
				row["Store Types"] = ", ".join(store_types)

				# Check PAM types
				
				has_pam = about['orchestrator'].get('pam_support', False)
				if has_pam:
					row["Feature Flags"] += "PAM;"
				else:
					row["Feature Flags"] += "needsPAM;"

			except Exception:
				logger.info(f"Warning: Unexpected store_types format in {repo_name}")
				logger.info("Please verify the manifest format for this repository.")
				# Continue with empty store types to avoid breaking the inventory generation
				row["Store Types"] = "(none)"
	
		# Get README last update
		date, sha = get_readme_last_commit(ORG, repo_name)
		if date and sha:
			row["Last README.md Update"] = f"{date} ({sha})"
		else:
			row["Last README.md Update"] = "No README.md or failed to fetch commit info"

		inventory.append(row)

	return inventory

def main():
	parser = argparse.ArgumentParser(
		description="Generate CSV inventory of Keyfactor integrations"
	)
	parser.add_argument(
		"--manifest-dir",
		help="Directory containing collected manifests. If not specified, fetches manifests live from GitHub"
	)
	parser.add_argument(
		"--output",
		required=True,
		help="Output CSV file path"
	)
	parser.add_argument(
		"--public-only",
		action="store_true",
		help="Only include public repositories"
	)
	parser.add_argument(
		"--integration-type",
		help="Filter by integration type (e.g., orchestrator, anyca, pam, dns-plugin)"
	)

	args = parser.parse_args()

	try:
		inventory = build_inventory(
			manifest_dir=args.manifest_dir,
			public_only=args.public_only,
			integration_type=args.integration_type
		)

		# Write CSV
		if inventory:
			fieldnames = [
				"Integration Name",
				"Integration Type",
				"Feature Flags",
				".NET Versions",
				"Store Types",
				"keyfactor/actions Version",
				"GitHub Repo URL",
				"Confluence Link",
				"Last README.md Update",
			]

			with open(args.output, 'w', newline='') as f:
				writer = csv.DictWriter(f, fieldnames=fieldnames)
				writer.writeheader()
				writer.writerows(inventory)

			print(f"Inventory written to {args.output} ({len(inventory)} entries)")
		else:
			print("No manifests found matching filter criteria")
			sys.exit(1)
	finally:
		cleanup_clones()

if __name__ == "__main__":
	main()
