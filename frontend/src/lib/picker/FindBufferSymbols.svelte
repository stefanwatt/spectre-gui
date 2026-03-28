<script lang="ts">
	import { FindBufferSymbols } from '@bindings/nvim-gui/features/picker/lspsymbolspicker.js';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindBufferSymbols(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-buffer-symbols'} title={'Find Buffer Symbols'} />
