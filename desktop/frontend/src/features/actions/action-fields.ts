import type { ModuleAction, OutputListItem, ResourceCatalog } from "../../app/types";

export interface FieldOption {
  label: string;
  value: string;
}

function toOptions(items: OutputListItem[]): FieldOption[] {
  return items
    .map((item) => ({
      label: item.secondary ? `${item.primary} · ${item.secondary}` : item.primary,
      value: item.value ?? item.primary,
    }))
    .filter((item, index, list) => item.value && list.findIndex((candidate) => candidate.value === item.value) === index);
}

export function buildActionValues(
  action: ModuleAction | null,
  seedValues: Record<string, string> = {},
  preferredDefaults: Record<string, string> = {}
): Record<string, string> {
  const out: Record<string, string> = { ...seedValues };
  for (const field of action?.fields ?? []) {
    if (!out[field.key]) {
      out[field.key] = preferredDefaults[field.key] || field.default || "";
    }
  }
  return out;
}

export function resolveActionFieldOptions(moduleID: string, actionID: string, fieldKey: string, catalog: ResourceCatalog): FieldOption[] {
  if (fieldKey === "method") {
    return [
      { label: "squash", value: "squash" },
      { label: "merge", value: "merge" },
      { label: "rebase", value: "rebase" },
    ];
  }
  if (fieldKey === "strategy") {
    return [
      { label: "ours（保留当前分支）", value: "ours" },
      { label: "theirs（保留对端版本）", value: "theirs" },
    ];
  }

  if (fieldKey === "remote") {
    return toOptions(catalog.remotes);
  }

  if (moduleID === "repo" && fieldKey === "path" && (actionID === "repo_stage_path" || actionID === "repo_unstage_path" || actionID === "repo_discard_path")) {
    return toOptions(catalog.repoStatus);
  }

  if (moduleID === "branch" && fieldKey === "name" && (actionID === "branch_checkout" || actionID === "branch_delete" || actionID === "branch_force_sync")) {
    return toOptions(
      actionID === "branch_delete"
        ? catalog.branches.filter((item) => !item.active)
        : catalog.branches
    );
  }

  if (moduleID === "worktree" && fieldKey === "base") {
    return toOptions(catalog.branches);
  }

  if (moduleID === "worktree" && fieldKey === "target") {
    if (actionID === "worktree_remove") {
      return toOptions(catalog.worktrees);
    }

    return [
      ...toOptions(catalog.issues),
      ...toOptions(catalog.branches.filter((item) => !item.active)),
    ];
  }

  if (moduleID === "issue" && (actionID === "issue_view" || actionID === "issue_close") && fieldKey === "id") {
    return toOptions(catalog.issues);
  }

  if (moduleID === "remote" && fieldKey === "name" && (actionID === "remote_rename" || actionID === "remote_update" || actionID === "remote_set_url" || actionID === "remote_set_push_url" || actionID === "remote_remove" || actionID === "remote_fetch")) {
    return toOptions(catalog.remotes);
  }

  if (moduleID === "tag" && (actionID === "tag_push" || actionID === "tag_force_sync") && fieldKey === "name") {
    return toOptions(catalog.tags);
  }

  if (moduleID === "tag" && actionID === "tag_delete" && fieldKey === "name") {
    return toOptions(catalog.tags);
  }

  if (moduleID === "tag" && actionID === "tag_delete" && fieldKey === "delete_remote") {
    return [
      { label: "仅删除本地", value: "false" },
      { label: "同时删除远端", value: "true" },
    ];
  }

  if (moduleID === "pr" && actionID === "pr_create" && fieldKey === "base") {
    return toOptions(catalog.branches);
  }

  if (moduleID === "pr" && (actionID === "pr_view" || actionID === "pr_close") && fieldKey === "id") {
    return toOptions(catalog.prs);
  }
  if (moduleID === "conflict" && (actionID === "conflict_open" || actionID === "conflict_resolve" || actionID === "conflict_mark_resolved") && fieldKey === "path") {
    return toOptions(catalog.conflicts);
  }

  return [];
}
