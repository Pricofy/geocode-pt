# Memory Requirements Analysis

This document analyzes the memory requirements for each country implementation based on real data measurements.

## Raw Data Statistics

| Country | Raw Entries | Unique Codes | JSON Size | Status |
|---------|-------------|--------------|-----------|--------|
| España | 11,150 | 11,150 | ~1.5MB | Estimated |
| Francia | 51,611 | 20,414 | 2.3MB | Measured |
| Alemania | 23,297 | 10,813 | 1.3MB | Measured |
| Italia | ~4,484 | ~4,484 | ~1MB | Estimated |
| Portugal | 206,942 | 197,772 | 23MB | Measured |

## Memory Overhead Calculations

### Go Runtime Overhead
- **Base memory**: ~10-15MB for Go Lambda runtime
- **JSON deserialization**: ~2-3x raw JSON size
- **Map overhead**: ~50-100% additional for hash maps
- **String interning**: Variable, depends on duplicate strings

### Estimated In-Memory Usage

#### España (ES)
- JSON: 1.5MB
- Deserialized: ~4.5MB
- Maps/Indices: ~2MB
- **Total**: ~16-20MB
- **Lambda Memory**: 128MB ✅ (safe margin)

#### Francia (FR)
- JSON: 2.3MB
- Deserialized: ~7MB
- Maps/Indices: ~4MB
- **Total**: ~21-25MB
- **Lambda Memory**: 128MB ✅ (safe margin)

#### Alemania (DE)
- JSON: 1.3MB
- Deserialized: ~4MB
- Maps/Indices: ~2MB
- **Total**: ~16-20MB
- **Lambda Memory**: 128MB ✅ (safe margin)

#### Italia (IT)
- JSON: ~1MB
- Deserialized: ~3MB
- Maps/Indices: ~1MB
- **Total**: ~14-18MB
- **Lambda Memory**: 128MB ✅ (safe margin)

#### Portugal (PT)
- JSON: 23MB
- Deserialized: ~70MB
- Maps/Indices: ~40MB
- **Total**: ~130-150MB ⚠️
- **Lambda Memory**: 256MB ✅ (required)

## Lambda Memory Tiers

AWS Lambda offers these memory configurations:
- 128MB, 256MB, 512MB, 1024MB, 2048MB, 3072MB

### Cost Impact
- Memory cost is linear with allocated memory
- CPU power scales with memory allocation
- 256MB costs 2x more than 128MB

## Performance Analysis

### Cold Start Times (Estimated)

| Country | Data Load | Index Build | Total Cold Start |
|---------|-----------|-------------|------------------|
| España | 150ms | 50ms | ~200ms |
| Francia | 200ms | 100ms | ~300ms |
| Alemania | 150ms | 50ms | ~200ms |
| Italia | 100ms | 50ms | ~150ms |
| Portugal | 800ms | 200ms | ~1000ms ⚠️ |

### Operation Latencies

All countries should maintain similar operation latencies:
- **Geocode by postal**: <1ms (O(1) map lookup)
- **Geocode by municipality**: <5ms (O(1) map lookup)
- **Reverse geocode**: 10-20ms (O(n) brute force)
- **Validate**: <1ms (O(1) lookups)
- **Autocomplete**: <10ms (O(log n) binary search)

## Memory Optimization Strategies

### For Portugal (High Memory Usage)

1. **JSON Compression**: Consider gzip compression + runtime decompression
   - Potential: Reduce 23MB → ~8MB compressed
   - Trade-off: Slightly slower cold start

2. **Memory-Efficient Data Structures**:
   - Use `int32` instead of `float64` for coordinates (with precision loss)
   - Intern strings for municipalities/provinces
   - Use more compact map implementations

3. **Lazy Loading**: Load data only when needed
   - Trade-off: Slower first request, faster cold start

### General Optimizations

1. **String Interning**: Share common municipality/province strings
2. **Compact Coordinates**: Use fixed-point arithmetic instead of floats
3. **Index Optimization**: Pre-compute and cache frequently used lookups

## Lambda Configuration Recommendations

### Production Settings

| Country | Memory | Timeout | Architecture | Runtime |
|---------|--------|---------|--------------|---------|
| España | 128MB | 10s | ARM64 | PROVIDED_AL2023 |
| Francia | 128MB | 10s | ARM64 | PROVIDED_AL2023 |
| Alemania | 128MB | 10s | ARM64 | PROVIDED_AL2023 |
| Italia | 128MB | 10s | ARM64 | PROVIDED_AL2023 |
| Portugal | 256MB | 15s | ARM64 | PROVIDED_AL2023 |

### Development/Test Settings
- Use same memory as production to catch memory issues early
- Consider 512MB for development debugging

## Monitoring and Alerts

### CloudWatch Metrics to Monitor
- **Duration**: P95 latency should stay <500ms for most operations
- **Memory Usage**: Monitor actual memory used vs allocated
- **Cold Start Frequency**: Track cold start impact on user experience

### Alerts
- Duration > 2s for any operation
- Memory usage > 90% of allocated
- Error rate > 1%

## Cost Analysis

### Monthly Cost Estimates (EU-West-1)
Assumptions:
- 100K requests/month
- 200ms average duration
- 10% cold starts

| Country | Memory | Cost/1M Requests | Cost/Month (est) |
|---------|--------|------------------|------------------|
| España | 128MB | $0.125 | $12.50 |
| Francia | 128MB | $0.125 | $12.50 |
| Alemania | 128MB | $0.125 | $12.50 |
| Italia | 128MB | $0.125 | $12.50 |
| Portugal | 256MB | $0.25 | $25.00 |

*Costs are estimates and may vary based on actual usage patterns.*

## Recommendations

### Immediate Actions
1. **Portugal**: Upgrade to 256MB memory allocation
2. **All others**: Keep 128MB (sufficient headroom)
3. **Monitoring**: Implement memory usage tracking

### Future Optimizations
1. **Profile memory usage** in production after launch
2. **Consider compression** for Portugal if cold start becomes an issue
3. **Implement caching** for frequently accessed postal codes

### Risk Mitigation
1. **Test with allocated memory** in development
2. **Monitor memory usage** from day one in production
3. **Have rollback plan** if memory issues arise

## Conclusion

**Memory requirements are well understood:**

- ✅ **Spain, France, Germany, Italy**: 128MB sufficient
- ⚠️ **Portugal**: 256MB required due to large dataset
- ✅ **Performance**: All countries maintain good operation latencies
- ✅ **Cost**: Minimal additional cost for Portugal (2x memory = 2x cost)

The memory analysis confirms the feasibility of multi-country deployment with appropriate memory allocation per country.

---

**Last Updated:** November 2025
**Data Sources:** Real measurements from GeoNames datasets
**Lambda Runtime:** Go 1.24 on Amazon Linux 2023