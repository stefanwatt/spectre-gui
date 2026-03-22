<script lang="ts">
	import { cursor } from '$lib/state.svelte';
	import DocumentationWindow from './DocumentationWindow.svelte';

	interface CompletionMenuProps {
		completion: App.CompletionState;
	}
	let { completion }: CompletionMenuProps = $props();

	// Kind icon map (nerd font icons)
	const kindIcons: Record<string, string> = {
		Text: '󰉿',
		Method: '󰊕',
		Function: '󰊕',
		Constructor: '󰒓',
		Field: '󰜢',
		Variable: '󰆦',
		Property: '󰖷',
		Class: '󱡠',
		Interface: '󱡠',
		Struct: '󱡠',
		Module: '󰅩',
		Unit: '󰪚',
		Value: '󰦨',
		Enum: '󰦨',
		EnumMember: '󰦨',
		Keyword: '󰻾',
		Constant: '󰏿',
		Snippet: '󱄽',
		Color: '󰏘',
		File: '󰈔',
		Reference: '󰬲',
		Folder: '󰉋',
		Event: '󱐋',
		Operator: '󰪚',
		TypeParameter: '󰬛'
	};

	// Position: find the DOM row element for the cursor line
	let domTop: number | null = $state(null);
	let menuRef: HTMLDivElement | undefined = $state(undefined);
	let flipAbove = $state(false);

	const lineHeight = 28; // 22px font + 3px padding top + 3px padding bottom
	const offsetLeft = 6; // line number gutter width in ch units

	$effect(() => {
		if (!completion.visible) {
			domTop = null;
			return;
		}
		const el = document.getElementById(`row-${cursor.row}`);
		if (el) {
			domTop = el.offsetTop + lineHeight; // position below the cursor line
		} else {
			domTop = null;
		}
	});

	// Check if menu should flip above cursor
	$effect(() => {
		if (domTop === null || !menuRef) return;
		const parentEl = menuRef.closest('.nvim-window');
		if (!parentEl) return;
		const parentRect = parentEl.getBoundingClientRect();
		const menuHeight = menuRef.offsetHeight;
		const spaceBelow = parentRect.bottom - (parentRect.top + domTop);
		flipAbove = menuHeight > spaceBelow;
	});

	let topPx = $derived.by(() => {
		if (domTop === null) return '0px';
		if (flipAbove) {
			return `${domTop - lineHeight - (menuRef?.offsetHeight ?? 0)}px`;
		}
		return `${domTop}px`;
	});

	let leftCh = $derived(`${completion.col + offsetLeft}ch`);

	// Scroll selected item into view
	$effect(() => {
		if (!menuRef || completion.selectedIndex < 1) return;
		const selectedEl = menuRef.querySelector('.completion-item.selected');
		if (selectedEl) {
			selectedEl.scrollIntoView({ block: 'nearest' });
		}
	});
</script>

{#if completion.visible && completion.items.length > 0}
	<div
		bind:this={menuRef}
		class="completion-menu absolute z-[210] overflow-y-auto rounded-md border border-surface0 bg-base-100 text-text drop-shadow-lg"
		style="top: {topPx}; left: {leftCh}; max-height: {10 * lineHeight}px;"
	>
		{#each completion.items as item, i (i)}
			<div
				class="completion-item flex items-center gap-2 px-2 whitespace-nowrap"
				class:selected={i + 1 === completion.selectedIndex}
				class:deprecated={item.deprecated}
			>
				<span
					class="kind-icon w-5 text-center shrink-0 kind-{item.kind.toLowerCase()}"
					>{kindIcons[item.kind] ?? '󰉿'}</span
				>
				<span class="label flex-1 truncate">{item.label}</span>
				{#if item.detail}
					<span class="detail text-xs opacity-50 truncate max-w-48">{item.detail}</span>
				{/if}
			</div>
		{/each}
	</div>
	<DocumentationWindow {menuRef} />
{/if}

<style>
	.completion-menu {
		font-size: 19px;
		line-height: 22px;
	}
	.completion-item {
		padding-top: 2px;
		padding-bottom: 2px;
		cursor: default;
	}
	.completion-item.selected {
		background: rgba(255, 255, 255, 0.1);
	}
	.completion-item.deprecated .label {
		text-decoration: line-through;
		opacity: 0.5;
	}
	/* Kind-specific colors */
	.kind-function,
	.kind-method,
	.kind-constructor {
		color: #dca561;
	}
	.kind-variable,
	.kind-field,
	.kind-property {
		color: #7dcfff;
	}
	.kind-class,
	.kind-interface,
	.kind-struct {
		color: #bb9af7;
	}
	.kind-keyword {
		color: #c0caf5;
	}
	.kind-snippet {
		color: #9ece6a;
	}
	.kind-text,
	.kind-value,
	.kind-constant,
	.kind-enum,
	.kind-enummember {
		color: #e0af68;
	}
	.kind-module,
	.kind-file,
	.kind-folder {
		color: #73daca;
	}
</style>
