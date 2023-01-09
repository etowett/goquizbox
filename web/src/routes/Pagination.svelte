<script>
  export let pagination;
</script>

<div class="pagination">
  <nav aria-label="Page navigation example">
    <ul class="pagination">
      {#if pagination.prev_page !== null}
        <li class="page-item"><a class="page-link" href="?page={pagination.prev_page}">Previous</a></li>
      {:else}
        <li class="page-item disabled">
          <a class="page-link" href={null} tabindex="-1" aria-disabled="true">Previous</a>
        </li>
      {/if}
      <li class="page-item"><a class="page-link" href={null}>Page {pagination.page} of {pagination.num_pages}</a></li>
      {#if pagination.prev_page !== null}
        <li class="page-item"><a class="page-link" href="?page={pagination.next_page}">Next</a></li>
      {:else}
      <li class="page-item disabled">
          <a class="page-link" href={null} tabindex="-1" aria-disabled="true">Next</a>
      </li>
      {/if}
    </ul>
  </nav>
</div>
