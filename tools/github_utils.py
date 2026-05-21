"""Shared utilities for GitHub API interactions."""

import os
import sys
import requests
import re
import xml.etree.ElementTree as ET
import subprocess
import tempfile
import shutil
from pathlib import Path
from tqdm import tqdm
import datetime 

GITHUB_API = "https://api.github.com"
ORG = "keyfactor"
CLONE_TEMP_DIR = None


import logging
class TqdmLoggingHandler(logging.Handler):
	def emit(self, record):
		try:
			msg = self.format(record)
			tqdm.write(msg)
		except Exception:
			self.handleError(record)

logging.basicConfig(level=logging.INFO, handlers=[TqdmLoggingHandler()])
logger = logging.getLogger(__name__)

def get_headers():
	"""Get headers with authentication token."""
	token = os.environ.get("GITHUB_TOKEN")
	if not token:
		logger.error("GITHUB_TOKEN environment variable not set.")
		sys.exit(1)
	return {
		"Authorization": f"token {token}",
		"Accept": "application/vnd.github+json"
	}


def _init_clone_dir():
	"""Initialize a temporary directory for clones."""
	global CLONE_TEMP_DIR
	if CLONE_TEMP_DIR is None:
		CLONE_TEMP_DIR = tempfile.mkdtemp(prefix="github_repos_")
		logger.info(f"Using temp directory for clones: {CLONE_TEMP_DIR}")
	return CLONE_TEMP_DIR


def cleanup_clones():
	"""Clean up temporary clone directory."""
	global CLONE_TEMP_DIR
	if CLONE_TEMP_DIR and os.path.exists(CLONE_TEMP_DIR):
		try:
			shutil.rmtree(CLONE_TEMP_DIR)
			logger.info(f"Cleaned up temporary directory: {CLONE_TEMP_DIR}")
		except Exception as e:
			logger.warning(f"Failed to clean up temp directory: {e}")
		CLONE_TEMP_DIR = None


def _get_repo_clone_path(owner, repo):
	"""Get the path where a repo should be cloned."""
	temp_dir = _init_clone_dir()
	return os.path.join(temp_dir, f"{owner}_{repo}")


def _clone_repo(owner, repo):
	"""Clone a repository into the temp directory (shallow clone for speed)."""
	clone_path = _get_repo_clone_path(owner, repo)

	if os.path.exists(clone_path):
		return clone_path

	url = f"https://github.com/{owner}/{repo}.git"
	try:
		subprocess.run(
			["git", "clone", "--depth", "1", url, clone_path],
			capture_output=True,
			check=True,
			timeout=60
		)
		return clone_path
	except subprocess.CalledProcessError as e:
		logger.error(f"Failed to clone {owner}/{repo}: {e.stderr.decode() if e.stderr else e}")
		return None
	except subprocess.TimeoutExpired:
		logger.error(f"Timeout cloning {owner}/{repo}")
		return None


def get_repos(org, public_only=False):
	"""Fetch all repositories from an organization."""
	repos = []
	page = 1
	while True:
		url = f"{GITHUB_API}/orgs/{org}/repos?per_page=100&page={page}"
		if public_only:
			url += "&type=public"
		resp = requests.get(url, headers=get_headers())
		if resp.status_code != 200:
			logger.error(f"Failed to get repos: {resp.status_code}")
			break
		data = resp.json()
		if not data:
			break
		repos.extend(data)
		page += 1
	return repos


def get_file_content(owner, repo, path):
	"""Fetch file content from a repository using git clone."""
	clone_path = _clone_repo(owner, repo)
	if not clone_path:
		return None

	file_path = os.path.join(clone_path, path)
	try:
		with open(file_path, 'r') as f:
			return f.read()
	except FileNotFoundError:
		return None
	except Exception as e:
		logger.error(f"Failed to read {path} from {owner}/{repo}: {e}")
		return None


def search_files_in_repo(owner, repo, pattern, path=""):
	"""Search for files matching a pattern in a repository by walking the filesystem."""
	clone_path = _clone_repo(owner, repo)
	if not clone_path:
		return []

	results = []
	search_root = os.path.join(clone_path, path) if path else clone_path

	try:
		for root, dirs, files in os.walk(search_root):
			# Skip .git directory
			dirs[:] = [d for d in dirs if d != '.git']

			for file in files:
				if pattern.lstrip('.') in file or file.endswith(pattern):
					file_path = os.path.join(root, file)
					rel_path = os.path.relpath(file_path, clone_path)
					results.append({
						'path': rel_path.replace(os.sep, '/'),
						'name': file
					})
	except Exception as e:
		logger.error(f"Failed to search files in {owner}/{repo}: {e}")

	return results


def is_public_repo(owner, repo):
	"""Check if a repository is public."""
	url = f"{GITHUB_API}/repos/{owner}/{repo}"
	resp = requests.get(url, headers=get_headers())
	if resp.status_code == 200:
		return not resp.json().get('private', True)
	return False


def get_csproj_versions(owner, repo, csproj_path=None):
	"""Extract .NET versions from csproj file(s).

	First tries to use the specified csproj_path if provided.
	Falls back to searching for .csproj files in the repo if no path specified or if not found.

	Args:
		owner: Repository owner
		repo: Repository name
		csproj_path: Optional path to a specific .csproj file (e.g., "ProjectName/ProjectName.csproj")

	Returns:
		List of .NET target framework versions found in csproj file(s)
	"""
	versions = set()

	# Try specific path first if provided
	if csproj_path:
		logger.info(f"Attempting to fetch .csproj from specified path '{csproj_path}' in {owner}/{repo}")
		content = get_file_content(owner, repo, csproj_path)
		if content:
			try:
				versions.update(__csproj_versions_from_content(content))
				return sorted(list(versions))
			except Exception:
				logger.info(f"Warning: Failed to extract .NET versions from specified csproj path '{csproj_path}' in {owner}/{repo}")
		else:
			logger.info(f"Warning: Specified csproj path '{csproj_path}' not found in {owner}/{repo}. Falling back to search.")
			# Continue to fallback search if specific path not found
			return get_csproj_versions(owner, repo, csproj_path=None)

	logger.debug(f"Searching for .csproj files in {owner}/{repo} to extract .NET versions")
	# Fallback: search for .csproj files in the repo
	results = search_files_in_repo(owner, repo, ".csproj")

	for item in results:
		logger.debug(f"Found .csproj file at path '{item['path']}' in {owner}/{repo}. Attempting to extract .NET versions.")
		path = item['path']
		content = get_file_content(owner, repo, path)
		if content:
			try:
				versions.update(__csproj_versions_from_content(content))
			except ET.ParseError:
				logger.warning(f"Failed to parse csproj XML for {owner}/{repo} at path '{path}'")
				pass
		else:
			logger.warning(f"Failed to fetch content for csproj file at path '{path}' in {owner}/{repo}")
			return ["(not a .NET project)"]
	return sorted(list(versions))


def __csproj_versions_from_content(content):
	"""Helper to extract .NET versions from csproj XML content."""
	versions = set()
	try:
		root = ET.fromstring(content)
		pgs = filter(lambda e: e.tag.endswith('PropertyGroup'), root)
		for prop_group in pgs:
			
			target_fw = prop_group.find('{*}TargetFramework')
			target_fws = prop_group.find('{*}TargetFrameworks')
			legacy_target_fw = prop_group.find('{*}TargetFrameworkVersion')

			if target_fw is not None and target_fw.text:
				logger.debug(f"Found TargetFramework: {target_fw.text.strip()}")
				versions.add(target_fw.text.strip())
			elif target_fws is not None and target_fws.text:
				logger.debug(f"Found TargetFrameworks: {target_fws.text.strip()}")
				for fw in target_fws.text.split(';'):
					versions.add(fw.strip())
			elif legacy_target_fw is not None and legacy_target_fw.text:
				logger.debug(f"Found legacy TargetFrameworkVersion: {legacy_target_fw.text.strip()}")
				versions.add(legacy_target_fw.text.strip() + " (legacy)")
	except ET.ParseError:
		logger.warning("Warning: Failed to parse csproj XML content.")
	
	logger.debug(f"Extracted .NET versions from csproj content: {', '.join(versions) if versions else 'None'}")
	return sorted(list(versions))

def get_workflow_actions_version(owner, repo):
	"""Find keyfactor/actions version in workflow files."""
	clone_path = _clone_repo(owner, repo)
	if not clone_path:
		return "(clone failed)"

	workflows_dir = os.path.join(clone_path, ".github", "workflows")
	if not os.path.exists(workflows_dir):
		return "(no workflows)"

	try:
		for file in os.listdir(workflows_dir):
			if file.endswith('.yml') or file.endswith('.yaml'):
				file_path = os.path.join(workflows_dir, file)
				with open(file_path, 'r') as f:
					logger.debug(f"Checking workflow file '{file}' for keyfactor/actions version in {owner}/{repo}")
					content = f.read()
					match = re.search(r'keyfactor/actions/.github/workflows/starter.yml@(v[\d.]+|[\d.]+)', content)
					logger.debug(f"Regex search for keyfactor/actions version in {owner}/{repo} workflow file '{file}' returned: {'Found version ' + match.group(1) if match else 'No match found'}")
					if match:
						return match.group(1)
					logger.debug(f"No keyfactor/actions version found in workflow file '{file}' for {owner}/{repo}")


	except Exception as e:
		logger.error(f"Failed to check workflows in {owner}/{repo}: {e}")

	return "(unknown/custom)"


def get_readme_last_commit(owner, repo):
	"""Get last commit date and SHA for README.md using git."""
	clone_path = _clone_repo(owner, repo)
	if not clone_path:
		return None, None

	readme_path = os.path.join(clone_path, "README.md")
	if not os.path.exists(readme_path):
		return None, None

	try:
		result = subprocess.run(
			["git", "log", "-1", "--format=%ai|%h", "README.md"],
			cwd=clone_path,
			capture_output=True,
			text=True,
			timeout=10
		)
		if result.returncode == 0 and result.stdout.strip():
			date, sha = result.stdout.strip().split('|')

			pdate  = datetime.datetime.fromisoformat(date).astimezone(datetime.timezone.utc) 

			return pdate, sha
	except Exception as e:
		logger.error(f"Failed to get README.md last commit for {owner}/{repo}: {e}")

	return None, None
