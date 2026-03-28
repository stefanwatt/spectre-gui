<script lang="ts">
	import { FindHelp } from '@bindings/nvim-gui/features/picker/picker';
	import Picker from './Picker.svelte';
	import { nestedState } from '$lib/state.svelte';

	function onQueryChanged(query: string) {
		FindHelp(query).then((updatedResults) => {
			nestedState.pickerResults = updatedResults.filter((r): r is NonNullable<typeof r> => r != null);
		}).catch((err) => {
			console.error('FindHelp RPC failed', err);
		});
	}
</script>

<Picker {onQueryChanged} mode={'find-help'} title={'Find Help'} />
