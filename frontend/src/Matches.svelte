<script>
  /** @type {RipgrepMatch[]}*/
  export let matches = [];
</script>

{#each matches as match}
  {#if match.Path.length < 15 && match.Col.length < 5}
    <div class="snap-start flex justify-start p-2">
      <code class="language-go w-4/5">
        {match.MatchedLine}
      </code>
    </div>
  {/if}
{/each}
