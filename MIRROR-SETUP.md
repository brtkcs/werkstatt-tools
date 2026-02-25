# Gitea → GitHub + Codeberg push mirror

Gitea supports multiple push mirrors per repository.
This means one `git push` to Gitea automatically syncs to both GitHub and Codeberg.

## Setup

### 1. Create empty repos on GitHub and Codeberg

- GitHub: `https://github.com/brtkcs/werkstatt` (empty, no README)
- Codeberg: `https://codeberg.org/brtkcs/werkstatt` (empty, no README)

### 2. Create access tokens

**GitHub:**
- Settings → Developer settings → Personal access tokens → Fine-grained tokens
- Repository access: Only select repositories → werkstatt
- Permissions: Contents (Read and write)
- Copy token

**Codeberg:**
- Settings → Applications → Generate New Token
- Select: `repo` scope
- Copy token

### 3. Add push mirrors in Gitea

Go to your werkstatt repo in Gitea → Settings → Repository → Mirror Settings

**GitHub mirror:**
- Git Remote Repository URL: `https://brtkcs@github.com/brtkcs/werkstatt.git`
- Password: your GitHub token
- Mirror Direction: Push
- Sync when commits are pushed: ✓

**Codeberg mirror:**
- Git Remote Repository URL: `https://brtkcs@codeberg.org/brtkcs/werkstatt.git`
- Password: your Codeberg token
- Mirror Direction: Push
- Sync when commits are pushed: ✓

### 4. Test

```bash
cd werkstatt
git add .
git commit -m "test: mirror sync"
git push
```

Check all three:
- Gitea: your server URL
- GitHub: https://github.com/brtkcs/werkstatt
- Codeberg: https://codeberg.org/brtkcs/werkstatt

All three should show the same commit within seconds.

## Notes

- Mirror sync happens on every push automatically
- If sync fails, Gitea retries periodically
- Check Gitea → Settings → Repository → Mirror Settings for sync status
- Each repo needs its own mirror setup (werkstatt, dotfiles, etc.)
