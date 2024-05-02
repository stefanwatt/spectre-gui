<script>
  import { Search } from "../wailsjs/go/main/App";

  let matches = [];
  let search_term = "utils";
  let dir = "/home/stefan/Projects/nvim-float";
  $: {
    Search(search_term, dir).then((res) => {
      matches = res;
    });
  }
</script>

<div class="text-text bg-base h-full w-full min-h-screen min-screen p-4">
  <div class="flex">
    <input
      bind:value={search_term}
      type="text"
      placeholder="Search..."
      class="w-full input input-primary bg-base text-text border-text border rounded-md px-2 py-1"
    />
    <span class="mx-2"></span>
    <input
      bind:value={dir}
      type="text"
      placeholder="Search..."
      class="w-full input input-primary bg-base text-text border-text border rounded-md px-2 py-1"
    />
  </div>
  <div class="mt-8 border-text border rounded-md p-1">
    {#each matches as match}
      <div class="flex justify-between">
        <div class="mx-2">{match.Path}:{match.Row}:{match.Col}</div>
        <div class="mx-2">{match.MatchedLine}</div>
      </div>
    {/each}
  </div>
</div>
