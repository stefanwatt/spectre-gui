<script lang="ts">
	interface Position {
		top: string;
		left: string;
	}
	interface FloatingWindow {
		z_index: number;
		content: string;
	}
	interface FloatingWindowProps {
		position: Position;
		win: FloatingWindow;
	}
	let { position, win }: FloatingWindowProps = $props();
	function decode(message: string) {
    const hexValues = message.split(',');
    let result = '';
    
    for (const hex of hexValues) {
        const codePoint = parseInt(hex, 16);
        result += String.fromCodePoint(codePoint);
    }
    
    return result;
	}
	let content = $derived(decode(win.content));
</script>

<div
	class="absolute overflow-hidden rounded-md border border-surface0 bg-crust text-text whitespace-pre p-1"
	style="top: {position.top}; left: {position.left}; z-index: {win.z_index || 100};"
>
	{content}
</div>
