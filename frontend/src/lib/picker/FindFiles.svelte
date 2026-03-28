<script lang="ts">
	import { FindFiles } from '@bindings/nvim-gui/features/picker/picker';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	function onQueryChanged(query: string) {
		FindFiles(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults.filter((r): r is NonNullable<typeof r> => r != null);
		}).catch((err) => {
			console.error('FindFiles RPC failed', err);
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-files'} title={'Find Files'} />
