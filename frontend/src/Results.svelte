<script>
  import { onMount } from "svelte";
  import "./assets/prism.css";
  import ChevronDown from "./ChevronDown.svelte";
  import ChevronUp from "./ChevronUp.svelte";

  /*** @type {RipgrepResult} */
  export let results;
  let collapsed = false;
  onMount(async () => {
    await import("./assets/prism.js");
    setTimeout(() => {
      window.Prism.highlightAll(false);
    });
  });
  $: {
    if (results && window.Prism) {
      window.Prism.highlightAll(false);
    }
  }
</script>

{#if results && Object.keys(results)?.length}
  {#each Object.entries(results) as [path, matches], i}
    {#if !path.includes(":")}
      <div class="flex pt-2 snap-start">
        <div class="text-blue">
          {#if !collapsed}
            <ChevronDown></ChevronDown>
          {:else}
            <ChevronUp></ChevronUp>
          {/if}
        </div>

        <div>
          <div class="flex">
            <div class="font-bold text-blue px-2">{path}</div>
            <div
              class="w-8 flex justify-center items-center rounded-full bg-blue text-base"
            >
              {matches.length}
            </div>
          </div>
          {#each matches as match, j}
            {#if match.Path.length < 15 && match.Col.length < 5}
              <div class="snap-start flex justify-start p-2">
                <code class="language-go w-4/5">
                  {match.MatchedLine}
                </code>
              </div>
            {/if}
          {/each}
        </div>
      </div>
    {/if}
  {/each}
{/if}
