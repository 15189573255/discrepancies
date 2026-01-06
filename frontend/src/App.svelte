<script lang="ts">
  import { onMount } from 'svelte';
  import {
    SelectZipFile,
    SelectWorkDir,
    SelectOutputDir,
    Compare,
    GetTextDiff,
    ExportDiffs,
    CreateZip,
    GetConfig,
    GetZipRootFolder,
    GetExcludeRules,
    SetExcludeRules,
    AddExcludeRule,
    RemoveExcludeRule,
    ResetExcludeRules
  } from '../wailsjs/go/main/App.js';
  import { EventsOn } from '../wailsjs/runtime/runtime.js';

  // Types
  interface DiffItem {
    relPath: string;
    type: 'added' | 'modified' | 'deleted';
    selected: boolean;
    sourcePath: string;
  }

  interface CompareResult {
    items: DiffItem[];
    totalFiles: number;
    added: number;
    modified: number;
    deleted: number;
  }

  interface TextDiff {
    oldContent: string;
    newContent: string;
    lines: { type: string; content: string }[];
  }

  interface ProgressEvent {
    current: number;
    total: number;
    message: string;
  }

  interface ExcludeRule {
    pattern: string;
    type: 'glob' | 'regex';
    isDir: boolean;
    enabled: boolean;
    comment: string;
  }

  // State
  let zipPath = '';
  let workDir = '';
  let outputDir = '';
  let diffItems: DiffItem[] = [];
  let compareResult: CompareResult | null = null;
  let selectedItem: DiffItem | null = null;
  let textDiff: TextDiff | null = null;
  let isComparing = false;
  let isExporting = false;
  let progressMessage = '';
  let progressPercent = 0;
  let errorMessage = '';
  let successMessage = '';

  // Settings state
  let showSettings = false;
  let excludeRules: ExcludeRule[] = [];
  let newRule: ExcludeRule = { pattern: '', type: 'glob', isDir: false, enabled: true, comment: '' };
  let editingIndex: number | null = null;

  // Computed
  $: selectedCount = diffItems.filter(i => i.selected && i.type !== 'deleted').length;
  $: allSelected = diffItems.length > 0 && diffItems.every(item => item.selected);

  onMount(async () => {
    try {
      const config = await GetConfig();
      if (config.lastZipPath) zipPath = config.lastZipPath;
      if (config.lastWorkDir) workDir = config.lastWorkDir;
      if (config.lastOutputDir) outputDir = config.lastOutputDir;
    } catch (e) {
      console.error('Failed to load config:', e);
    }

    EventsOn('backend:progress', (event: ProgressEvent) => {
      progressMessage = event.message;
      progressPercent = event.total > 0 ? Math.round((event.current / event.total) * 100) : 0;
    });
  });

  async function selectZip() {
    try {
      const path = await SelectZipFile();
      if (path) {
        zipPath = path;
        clearResults();
      }
    } catch (e) {
      showError('选择 ZIP 文件失败: ' + e);
    }
  }

  async function selectWork() {
    try {
      const path = await SelectWorkDir();
      if (path) {
        workDir = path;
        clearResults();
      }
    } catch (e) {
      showError('选择工作目录失败: ' + e);
    }
  }

  async function selectOutput() {
    try {
      const path = await SelectOutputDir();
      if (path) {
        outputDir = path;
      }
    } catch (e) {
      showError('选择输出目录失败: ' + e);
    }
  }

  async function doCompare() {
    if (!zipPath || !workDir) {
      showError('请先选择 ZIP 文件和工作目录');
      return;
    }

    clearMessages();
    isComparing = true;
    progressMessage = '正在比较...';

    try {
      const result = await Compare(zipPath, workDir);
      compareResult = result;
      diffItems = result.items;
      progressMessage = '';

      if (diffItems.length === 0) {
        showSuccess('没有发现差异，两个目录内容相同');
      } else {
        showSuccess(`发现 ${diffItems.length} 个差异文件`);
      }
    } catch (e) {
      showError('比较失败: ' + e);
    } finally {
      isComparing = false;
    }
  }

  async function viewDiff(item: DiffItem) {
    selectedItem = item;
    textDiff = null;

    if (item.type === 'deleted') {
      return;
    }

    try {
      const diff = await GetTextDiff(zipPath, workDir, item.relPath);
      textDiff = diff;
    } catch (e) {
      // Non-text file or error
      textDiff = null;
    }
  }

  function toggleSelectAll() {
    const newValue = !allSelected;
    diffItems = diffItems.map(item => ({ ...item, selected: newValue }));
  }

  function toggleSelect(index: number) {
    diffItems[index].selected = !diffItems[index].selected;
    diffItems = [...diffItems];
  }

  async function doExport() {
    const selectedItems = diffItems.filter(item => item.selected && item.type !== 'deleted');
    if (selectedItems.length === 0) {
      showError('请选择要导出的文件');
      return;
    }

    if (!outputDir) {
      showError('请先选择输出目录');
      return;
    }

    clearMessages();
    isExporting = true;
    progressMessage = '正在导出...';

    try {
      await ExportDiffs(selectedItems, outputDir);
      showSuccess(`成功导出 ${selectedItems.length} 个文件`);
    } catch (e) {
      showError('导出失败: ' + e);
    } finally {
      isExporting = false;
      progressMessage = '';
    }
  }

  async function doExportAndZip() {
    const selectedItems = diffItems.filter(item => item.selected && item.type !== 'deleted');
    if (selectedItems.length === 0) {
      showError('请选择要导出的文件');
      return;
    }

    if (!outputDir) {
      showError('请先选择输出目录');
      return;
    }

    clearMessages();
    isExporting = true;
    progressMessage = '正在导出...';

    try {
      await ExportDiffs(selectedItems, outputDir);
      progressMessage = '正在创建 ZIP...';
      const rootFolder = await GetZipRootFolder(zipPath);
      const zipFilePath = await CreateZip(outputDir, rootFolder || 'output');
      showSuccess(`成功创建: ${zipFilePath}`);
    } catch (e) {
      showError('操作失败: ' + e);
    } finally {
      isExporting = false;
      progressMessage = '';
    }
  }

  function clearResults() {
    diffItems = [];
    compareResult = null;
    selectedItem = null;
    textDiff = null;
    clearMessages();
  }

  function showError(msg: string) {
    errorMessage = msg;
    successMessage = '';
  }

  function showSuccess(msg: string) {
    successMessage = msg;
    errorMessage = '';
  }

  function clearMessages() {
    errorMessage = '';
    successMessage = '';
  }

  function getTypeText(type: string): string {
    switch (type) {
      case 'added': return '新增';
      case 'modified': return '修改';
      case 'deleted': return '删除';
      default: return type;
    }
  }

  function getFileName(path: string): string {
    return path.split('/').pop() || path;
  }

  // Settings functions
  async function openSettings() {
    try {
      excludeRules = await GetExcludeRules();
      showSettings = true;
      resetNewRule();
    } catch (e) {
      showError('加载排除规则失败: ' + e);
    }
  }

  function closeSettings() {
    showSettings = false;
    editingIndex = null;
    resetNewRule();
  }

  function resetNewRule() {
    newRule = { pattern: '', type: 'glob', isDir: false, enabled: true, comment: '' };
    editingIndex = null;
  }

  async function saveRule() {
    if (!newRule.pattern.trim()) {
      showError('请输入匹配模式');
      return;
    }

    try {
      if (editingIndex !== null) {
        // Update existing rule
        excludeRules[editingIndex] = { ...newRule };
        await SetExcludeRules(excludeRules);
      } else {
        // Add new rule
        await AddExcludeRule(newRule);
        excludeRules = await GetExcludeRules();
      }
      resetNewRule();
    } catch (e) {
      showError('保存规则失败: ' + e);
    }
  }

  function editRule(index: number) {
    editingIndex = index;
    newRule = { ...excludeRules[index] };
  }

  function cancelEdit() {
    resetNewRule();
  }

  async function deleteRule(index: number) {
    try {
      await RemoveExcludeRule(index);
      excludeRules = await GetExcludeRules();
    } catch (e) {
      showError('删除规则失败: ' + e);
    }
  }

  async function toggleRuleEnabled(index: number) {
    try {
      excludeRules[index].enabled = !excludeRules[index].enabled;
      await SetExcludeRules(excludeRules);
    } catch (e) {
      showError('更新规则失败: ' + e);
    }
  }

  async function resetRules() {
    try {
      await ResetExcludeRules();
      excludeRules = await GetExcludeRules();
      showSuccess('已重置为默认规则');
    } catch (e) {
      showError('重置规则失败: ' + e);
    }
  }
</script>

<main class="h-screen flex flex-col bg-zinc-100 p-4 gap-4">
  <!-- Header - Path Selection -->
  <div class="card p-5">
    <div class="grid grid-cols-1 gap-4">
      <!-- ZIP File -->
      <div class="flex items-center gap-3">
        <label class="w-20 text-right text-sm font-medium text-zinc-700">原始 ZIP</label>
        <div class="flex-1 flex gap-2">
          <input
            type="text"
            class="input flex-1"
            readonly
            value={zipPath}
            placeholder="选择原始 ZIP 压缩包..."
          />
          <button class="btn btn-secondary whitespace-nowrap" on:click={selectZip}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            浏览
          </button>
        </div>
      </div>

      <!-- Work Directory -->
      <div class="flex items-center gap-3">
        <label class="w-20 text-right text-sm font-medium text-zinc-700">工作目录</label>
        <div class="flex-1 flex gap-2">
          <input
            type="text"
            class="input flex-1"
            readonly
            value={workDir}
            placeholder="选择工作目录..."
          />
          <button class="btn btn-secondary whitespace-nowrap" on:click={selectWork}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            浏览
          </button>
        </div>
      </div>

      <!-- Output Directory -->
      <div class="flex items-center gap-3">
        <label class="w-20 text-right text-sm font-medium text-zinc-700">输出目录</label>
        <div class="flex-1 flex gap-2">
          <input
            type="text"
            class="input flex-1"
            readonly
            value={outputDir}
            placeholder="选择输出目录..."
          />
          <button class="btn btn-secondary whitespace-nowrap" on:click={selectOutput}>
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z" />
            </svg>
            浏览
          </button>
        </div>
      </div>
    </div>

    <!-- Compare Button and Settings -->
    <div class="mt-5 pt-5 border-t border-zinc-200 flex justify-center gap-3">
      <button
        class="btn btn-primary px-8"
        on:click={doCompare}
        disabled={isComparing || !zipPath || !workDir}
      >
        {#if isComparing}
          <svg class="animate-spin h-4 w-4" fill="none" viewBox="0 0 24 24">
            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
            <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
          </svg>
          比较中...
        {:else}
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2" />
          </svg>
          开始比较
        {/if}
      </button>
      <button
        class="btn btn-secondary"
        on:click={openSettings}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
        </svg>
        排除规则
      </button>
    </div>
  </div>

  <!-- Messages -->
  {#if errorMessage}
    <div class="rounded-lg bg-red-50 p-4 ring-1 ring-inset ring-red-200">
      <div class="flex items-center gap-3">
        <svg class="h-5 w-5 text-red-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="text-sm text-red-700">{errorMessage}</p>
      </div>
    </div>
  {/if}

  {#if successMessage}
    <div class="rounded-lg bg-emerald-50 p-4 ring-1 ring-inset ring-emerald-200">
      <div class="flex items-center gap-3">
        <svg class="h-5 w-5 text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
        </svg>
        <p class="text-sm text-emerald-700">{successMessage}</p>
      </div>
    </div>
  {/if}

  {#if progressMessage}
    <div class="rounded-lg bg-zinc-100 p-4 ring-1 ring-inset ring-zinc-200">
      <div class="flex items-center gap-3">
        <svg class="animate-spin h-5 w-5 text-zinc-500" fill="none" viewBox="0 0 24 24">
          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
        </svg>
        <p class="text-sm text-zinc-600">{progressMessage}</p>
        {#if progressPercent > 0}
          <span class="text-sm text-zinc-500">({progressPercent}%)</span>
        {/if}
      </div>
    </div>
  {/if}

  <!-- Main Content -->
  <div class="flex-1 flex gap-4 min-h-0">
    <!-- Left: Diff List -->
    <div class="w-[420px] flex-shrink-0 card flex flex-col">
      <!-- Header -->
      <div class="px-5 py-4 border-b border-zinc-200">
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold text-zinc-900">差异列表</h2>
          {#if compareResult}
            <div class="flex items-center gap-2 text-xs text-zinc-500">
              <span class="inline-flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-emerald-500"></span>
                {compareResult.added}
              </span>
              <span class="inline-flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-amber-500"></span>
                {compareResult.modified}
              </span>
              <span class="inline-flex items-center gap-1">
                <span class="w-2 h-2 rounded-full bg-red-500"></span>
                {compareResult.deleted}
              </span>
            </div>
          {/if}
        </div>
      </div>

      {#if diffItems.length > 0}
        <!-- Toolbar -->
        <div class="px-5 py-3 border-b border-zinc-100 flex items-center justify-between">
          <label class="flex items-center gap-2 text-sm text-zinc-600 cursor-pointer">
            <input
              type="checkbox"
              class="checkbox"
              checked={allSelected}
              on:change={toggleSelectAll}
            />
            全选
          </label>
          <span class="text-xs text-zinc-400">
            已选 {selectedCount} 项
          </span>
        </div>

        <!-- List -->
        <div class="flex-1 overflow-y-auto">
          {#each diffItems as item, index}
            <div
              class="group flex items-center gap-3 px-5 py-3 border-b border-zinc-50 cursor-pointer transition-colors
                     hover:bg-zinc-50
                     {selectedItem === item ? 'bg-zinc-100' : ''}"
              on:click={() => viewDiff(item)}
              on:keypress={(e) => e.key === 'Enter' && viewDiff(item)}
              tabindex="0"
              role="button"
            >
              <input
                type="checkbox"
                class="checkbox"
                checked={item.selected}
                on:click|stopPropagation={() => toggleSelect(index)}
              />
              <span class="tag tag-{item.type}">{getTypeText(item.type)}</span>
              <div class="flex-1 min-w-0">
                <p class="text-sm font-medium text-zinc-900 truncate">{getFileName(item.relPath)}</p>
                <p class="text-xs text-zinc-500 truncate">{item.relPath}</p>
              </div>
              <svg class="w-4 h-4 text-zinc-400 opacity-0 group-hover:opacity-100 transition-opacity" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
              </svg>
            </div>
          {/each}
        </div>
      {:else}
        <div class="flex-1 flex items-center justify-center">
          <div class="text-center">
            <svg class="mx-auto h-12 w-12 text-zinc-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
            <p class="mt-2 text-sm text-zinc-500">
              {#if isComparing}
                正在比较...
              {:else}
                选择文件后点击"开始比较"
              {/if}
            </p>
          </div>
        </div>
      {/if}
    </div>

    <!-- Right: Diff Preview -->
    <div class="flex-1 card flex flex-col min-w-0">
      <!-- Header -->
      <div class="px-5 py-4 border-b border-zinc-200">
        <div class="flex items-center justify-between">
          <h2 class="text-base font-semibold text-zinc-900">差异预览</h2>
          {#if selectedItem}
            <span class="text-xs text-zinc-500 truncate max-w-xs">{selectedItem.relPath}</span>
          {/if}
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-auto">
        {#if textDiff && textDiff.lines.length > 0}
          <div class="font-mono text-xs leading-relaxed">
            {#each textDiff.lines as line, i}
              <div class="flex hover:bg-zinc-50
                          {line.type === 'insert' ? 'bg-emerald-50' : ''}
                          {line.type === 'delete' ? 'bg-red-50' : ''}">
                <span class="w-12 px-2 py-0.5 text-right text-zinc-400 select-none border-r border-zinc-200 flex-shrink-0">
                  {i + 1}
                </span>
                <span class="w-6 py-0.5 text-center flex-shrink-0
                            {line.type === 'insert' ? 'text-emerald-600 bg-emerald-100' : ''}
                            {line.type === 'delete' ? 'text-red-600 bg-red-100' : 'text-zinc-400'}">
                  {#if line.type === 'insert'}+{:else if line.type === 'delete'}-{:else}&nbsp;{/if}
                </span>
                <span class="flex-1 px-3 py-0.5 whitespace-pre-wrap break-all
                            {line.type === 'insert' ? 'text-emerald-900' : ''}
                            {line.type === 'delete' ? 'text-red-900' : 'text-zinc-700'}">{line.content || ' '}</span>
              </div>
            {/each}
          </div>
        {:else if selectedItem}
          <div class="flex-1 flex items-center justify-center h-full">
            <div class="text-center py-12">
              <svg class="mx-auto h-12 w-12 text-zinc-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              <p class="mt-2 text-sm text-zinc-500">
                {#if selectedItem.type === 'deleted'}
                  已删除的文件无法预览
                {:else if selectedItem.type === 'added'}
                  新增文件（二进制或无法预览）
                {:else}
                  无法预览此文件类型
                {/if}
              </p>
            </div>
          </div>
        {:else}
          <div class="flex-1 flex items-center justify-center h-full">
            <div class="text-center py-12">
              <svg class="mx-auto h-12 w-12 text-zinc-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
              </svg>
              <p class="mt-2 text-sm text-zinc-500">选择左侧文件查看差异</p>
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <!-- Footer -->
  <div class="card p-4 flex items-center justify-between">
    <div class="text-sm text-zinc-500">
      {#if selectedCount > 0}
        已选择 {selectedCount} 个文件
      {:else}
        未选择文件
      {/if}
    </div>
    <div class="flex gap-3">
      <button
        class="btn btn-secondary"
        on:click={doExport}
        disabled={isExporting || selectedCount === 0}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12" />
        </svg>
        导出选中项
      </button>
      <button
        class="btn btn-primary"
        on:click={doExportAndZip}
        disabled={isExporting || selectedCount === 0}
      >
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 8h14M5 8a2 2 0 110-4h14a2 2 0 110 4M5 8v10a2 2 0 002 2h10a2 2 0 002-2V8m-9 4h4" />
        </svg>
        导出并打包
      </button>
    </div>
  </div>

  <!-- Settings Modal -->
  {#if showSettings}
    <div class="fixed inset-0 z-50 overflow-y-auto">
      <!-- Backdrop -->
      <div class="fixed inset-0 bg-zinc-900/50 backdrop-blur-sm" on:click={closeSettings} on:keypress={(e) => e.key === 'Escape' && closeSettings()} tabindex="-1" role="button"></div>

      <!-- Modal -->
      <div class="relative min-h-screen flex items-center justify-center p-4">
        <div class="relative bg-white rounded-xl shadow-xl w-full max-w-3xl max-h-[80vh] flex flex-col">
          <!-- Header -->
          <div class="flex items-center justify-between px-6 py-4 border-b border-zinc-200">
            <h2 class="text-lg font-semibold text-zinc-900">排除规则设置</h2>
            <button
              class="p-1 text-zinc-400 hover:text-zinc-600 transition-colors"
              on:click={closeSettings}
            >
              <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- Content -->
          <div class="flex-1 overflow-y-auto p-6">
            <!-- Add/Edit Rule Form -->
            <div class="mb-6 p-4 bg-zinc-50 rounded-lg ring-1 ring-zinc-200">
              <h3 class="text-sm font-medium text-zinc-700 mb-3">
                {editingIndex !== null ? '编辑规则' : '添加新规则'}
              </h3>
              <div class="grid grid-cols-1 gap-3">
                <div class="flex gap-3">
                  <input
                    type="text"
                    class="input flex-1"
                    placeholder="匹配模式，如 *.log, node_modules, ^test.*"
                    bind:value={newRule.pattern}
                  />
                  <select class="input w-28" bind:value={newRule.type}>
                    <option value="glob">通配符</option>
                    <option value="regex">正则</option>
                  </select>
                </div>
                <div class="flex gap-3">
                  <input
                    type="text"
                    class="input flex-1"
                    placeholder="备注说明（可选）"
                    bind:value={newRule.comment}
                  />
                  <label class="flex items-center gap-2 text-sm text-zinc-600 whitespace-nowrap">
                    <input type="checkbox" class="checkbox" bind:checked={newRule.isDir} />
                    仅目录
                  </label>
                </div>
                <div class="flex justify-end gap-2">
                  {#if editingIndex !== null}
                    <button class="btn btn-secondary" on:click={cancelEdit}>取消</button>
                  {/if}
                  <button class="btn btn-primary" on:click={saveRule}>
                    {editingIndex !== null ? '保存修改' : '添加规则'}
                  </button>
                </div>
              </div>
            </div>

            <!-- Rules List -->
            <div class="space-y-2">
              <div class="flex items-center justify-between mb-3">
                <h3 class="text-sm font-medium text-zinc-700">已配置规则</h3>
                <button class="text-sm text-zinc-500 hover:text-zinc-700" on:click={resetRules}>
                  重置为默认
                </button>
              </div>

              {#if excludeRules.length === 0}
                <div class="text-center py-8 text-zinc-500">
                  暂无排除规则
                </div>
              {:else}
                {#each excludeRules as rule, index}
                  <div class="flex items-center gap-3 p-3 rounded-lg ring-1 ring-zinc-200 hover:ring-zinc-300 transition-colors
                              {!rule.enabled ? 'opacity-50 bg-zinc-50' : 'bg-white'}">
                    <input
                      type="checkbox"
                      class="checkbox"
                      checked={rule.enabled}
                      on:change={() => toggleRuleEnabled(index)}
                    />
                    <div class="flex-1 min-w-0">
                      <div class="flex items-center gap-2">
                        <code class="text-sm font-mono text-zinc-900 truncate">{rule.pattern}</code>
                        <span class="tag {rule.type === 'regex' ? 'bg-purple-50 text-purple-700 ring-purple-600/20' : 'bg-blue-50 text-blue-700 ring-blue-600/20'}">
                          {rule.type === 'regex' ? '正则' : '通配符'}
                        </span>
                        {#if rule.isDir}
                          <span class="tag bg-zinc-100 text-zinc-600 ring-zinc-200">目录</span>
                        {/if}
                      </div>
                      {#if rule.comment}
                        <p class="text-xs text-zinc-500 mt-0.5 truncate">{rule.comment}</p>
                      {/if}
                    </div>
                    <div class="flex items-center gap-1">
                      <button
                        class="p-1.5 text-zinc-400 hover:text-zinc-600 transition-colors"
                        on:click={() => editRule(index)}
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
                        </svg>
                      </button>
                      <button
                        class="p-1.5 text-zinc-400 hover:text-red-500 transition-colors"
                        on:click={() => deleteRule(index)}
                      >
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                      </button>
                    </div>
                  </div>
                {/each}
              {/if}
            </div>

            <!-- Help Text -->
            <div class="mt-6 p-4 bg-amber-50 rounded-lg ring-1 ring-amber-200">
              <h4 class="text-sm font-medium text-amber-800 mb-2">匹配规则说明</h4>
              <ul class="text-xs text-amber-700 space-y-1">
                <li><code class="bg-amber-100 px-1 rounded">*</code> 匹配任意字符（不含路径分隔符）</li>
                <li><code class="bg-amber-100 px-1 rounded">**</code> 匹配任意路径（含子目录）</li>
                <li><code class="bg-amber-100 px-1 rounded">?</code> 匹配单个字符</li>
                <li>正则模式使用标准正则表达式语法</li>
                <li>勾选"仅目录"时只匹配目录名，否则匹配文件名或完整路径</li>
              </ul>
            </div>
          </div>

          <!-- Footer -->
          <div class="flex justify-end px-6 py-4 border-t border-zinc-200">
            <button class="btn btn-primary" on:click={closeSettings}>
              完成
            </button>
          </div>
        </div>
      </div>
    </div>
  {/if}
</main>
