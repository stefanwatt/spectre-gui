<script>
  import { onMount } from "svelte";

  export let highlighted = "",
    text = "";
  let start = "",
    end = "",
    highlighted_start_index;

  $: {
    highlight(highlighted, text);
  }

  /*** @param {string} highlighted
   *    @param {string} text
   */
  function highlight(highlighted, text) {
    highlighted_start_index = text.indexOf(highlighted);
    start = text.slice(0, highlighted_start_index);
    end = text.slice(highlighted_start_index + highlighted.length);
  }
</script>

<div>
  <span>{start}</span>
  <span class="font-bold text-white bg-red-500">{highlighted}</span>
  <span>{end}</span>
</div>
