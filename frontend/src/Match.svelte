<script>
  /** @type {RipgrepMatch}*/
  export let match;

  /** @param {RipgrepMatch} match*/
  function replace_match(match) {
    console.log("replace_match", match);
  }
</script>

<button
  on:click={() => {
    replace_match(match);
  }}
  class="cursor-pointer snap-start flex justify-start p-2 w-full"
>
  <code class="language-go w-4/5">
    {match.MatchedLine}
  </code>
</button>
