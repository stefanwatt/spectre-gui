<script lang="ts">
	import { FindBufferSymbols } from '@bindings/nvim-gui/features/picker/lspsymbolspicker';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	function onQueryChanged(query: string) {
		FindBufferSymbols(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults.filter((r): r is NonNullable<typeof r> => r != null);
		}).catch((err) => {
			console.error('FindBufferSymbols RPC failed', err);
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-buffer-symbols'} title={'Find Buffer Symbols'} />
