<script>
  export let path = "",
    match_count = 0;
</script>

<div class="flex">
  <div class="font-bold text-blue px-2">{path}</div>
  <div
    class="w-8 flex justify-center items-center rounded-full bg-blue text-base font-mono"
  >
    {match_count}
  </div>
</div>
