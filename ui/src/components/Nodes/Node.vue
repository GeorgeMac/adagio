<template>
  <div class="container" v-if="this.node != null">
    <div class="column has-text-centered has-bottom">
      <p class="subtitle">Node Specification</p>
      <b-table
        ref="table"
        :data="data"
        detailed
        hoverable
        custom-detail-row
        detail-key="property"
        :opened-detailed="['retry']"
        :show-detail-icon="false">

        <template slot-scope="props">
          <b-table-column field="property" label="Property" width="250"> {{ props.row.property }} </b-table-column>
          <b-table-column field="value" label="Value">
            <template v-if="isRetry(props.row)">
              <a @click="toggle(props.row)">
                <b-icon pack="fas" icon="angle-right" :custom-class="retryClass()"></b-icon>
              </a>
            </template>
            <template v-else>
              {{ props.row.value }}
            </template>
          </b-table-column>
        </template>

        <template slot="detail" slot-scope="props">
          <tr class="" v-for="(result) in props.row.value" :key="result[0]">
            <td class="has-text-danger">&nbsp;&nbsp;&nbsp;{{ result[0] }}</td>
            <td>up to {{ result[1].max_attempts }} time(s)</td>
          </tr>
        </template>
      </b-table>
    </div>
    <div class="column has-text-centered has-bottom">
      <p class="subtitle">Node Execution</p>
      <b-table ref="attempts-table" :data="attempts">
        <template slot-scope="props">
          <b-table-column field="conclusion" label="Conclusion" width="250">{{ props.row.conclusion.toLowerCase() }}</b-table-column>
          <b-table-column field="output" label="Output">{{ decode(props.row.output) }}</b-table-column>
        </template>
      </b-table>
    </div>
  </div>
  <div v-else>Please select a node to see further details...</div>
</template>

<script>
export default {
  name: 'Node',
  props: {
    node: Object
  },
  data() {
    return {
      isRetryExpanded: false
    }
  },
  computed: {
    data() {
      if (this.node === undefined || this.node == null) {
        return []; 
      }

      var metadata = this.node.spec.metadata || {};
      metadata = Object.entries(metadata).map((v) => {
        var values = v[1].values
        return { 'property': v[0], 'value': values ? values.join(", ") : '' };
      });

      var retry = this.node.spec.retry || {};
      retry = {
        'property': 'retry',
        'value':    Object.keys(retry).length > 0 ? Object.entries(retry) : ''
      };

      return [
        { 'property': 'name', 'value': this.node.spec.name },
        { 'property': 'runtime', 'value': this.node.spec.runtime },
      ].concat(metadata, retry)
    },
    attempts() {
      return this.attemptsReversed()
    }
  },
  methods: {
    attemptsReversed() {
      if (this.node.attempts === undefined) {
        return [];
      }

      return this.node.attempts.slice().reverse()
    },
    toggle(row) {
      this.isRetryExpanded = !this.isRetryExpanded;
      this.$refs.table.toggleDetails(row)
    },
    isRetry(row) {
      return row.property == 'retry' && row.value.length > 0;
    },
    retryClass() {
      return this.isRetryExpanded ? 'rotate down' : 'rotate';
    },
    decode(output) {
      return output ? atob(output) : '';
    }
  }
}
</script>

<style>
.has-gutter {
  margin-bottom: 1rem;
}

.has-bottom {
  padding-bottom: 1rem;
}

.rotate{
    -moz-transition: all 0.1s linear;
    -webkit-transition: all 0.1s linear;
    transition: all 0.1s linear;
}

.rotate.down{
    -ms-transform: rotate(90deg);
    -moz-transform: rotate(90deg);
    -webkit-transform: rotate(90deg);
    transform: rotate(90deg);
}
</style>
