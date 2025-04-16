<script lang="ts">
	import { FindBufferSymbols } from '$lib/wailsjs/go/main/App';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindBufferSymbols(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker
	{onQueryChanged}
	mode={'find-buffer-symbols'}
	pickerPropName={'findBufferSymbols'}
	title={'Find Buffer Symbols'}
/>
