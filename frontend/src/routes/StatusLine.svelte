<script>
	/**@type{App.VimMode}*/
	export let mode = 'normal';
	export let cursor ={ row: 0, col: 0 }
	const mode_names = new Map();
	mode_names.set('normal', 'normal');
	mode_names.set('insert', 'insert');
	mode_names.set('visual', 'visual');
	mode_names.set('V', 'v-line');
	mode_names.set('cmdline_normal', 'command');
	mode_names.set('cmdline_insert', 'command');
	$: mode_name = mode_names.get(mode);
</script>

<div class="flex h-full w-full items-center bg-crust p-1 text-text">
	<span
		class:bg-blue={mode === 'normal'}
		class:bg-green={mode === 'insert'}
		class:bg-mauve={mode === 'visual' }
		class:bg-peach={mode === 'cmdline_normal'}
		class="rounded-md px-2 text-xl font-bold uppercase text-mantle"
	>
		{mode_name}
	</span>
	<span class="flex-grow">
	</span>
	<span 
		class="w-24 flex items-center justify-center rounded-md bg-blue px-2 text-xl font-bold uppercase text-mantle"
	>
		{cursor.row+1}:{cursor.col+1}
	</span>
</div>
