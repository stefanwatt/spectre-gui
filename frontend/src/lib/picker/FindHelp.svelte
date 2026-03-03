<script lang="ts">
	import { FindHelp } from '@bindings/nvim-gui/picker/picker.js';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	$effect(() => {});

	function onQueryChanged(query: string) {
		FindHelp(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults;
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-help'} title={'Find Help'} />
