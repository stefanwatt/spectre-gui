<script>
  import {Ping} from "$lib/wailsjs/go/main/App"

	let response = '';
	let last_time = 0;
  let start_time,end_time;
  window.addEventListener('keydown', async(e)=>{

    start_time = performance.now();
    response = await Ping(e.key)
    end_time = performance.now()
    console.log(`round trip: start_time ${start_time} ; end_time ${end_time}`)
    last_time = end_time-start_time;
  })
</script>

<div>Ping</div>

<div>
	response: {response}
	{#if last_time !== 0}
		took {last_time} ms
	{/if}
</div>
