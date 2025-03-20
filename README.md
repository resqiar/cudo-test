# Readme

- Original commit when Live Test was done: [8e4918f9daff115775cb01ae50b9b867ca3a8976](https://github.com/resqiar/cudo-test/commit/8e4918f9daff115775cb01ae50b9b867ca3a8976) 
- Complete commit done outside of Live Test: [08bc98b8875c4a873fd12ceba78eb6a3340c9a9e](https://github.com/resqiar/cudo-test/commit/08bc98b8875c4a873fd12ceba78eb6a3340c9a9e)

### Endpoints

```bash
GET http://localhost:9090/api/v1/fraud-detection?risk_level=Medium&limit=10000
```

Available query params:
1. limit -> limit lookup for recent tx, default to last 1000
1. risk_level -> only show risk level mentioned; can be separated by comma to show multiple; default to low;


### Response

```json
{
  "transactions": [
    {
      "transaction_id": "b51e5b85-ad9d-3dda-a968-60bd04a9afe4",
      "fraud_score": 55.827768165995735,
      "risk_level": "medium",
      "detection_results": {
        "amount_check": {
          "confidence_score": 100,
          "is_suspicious": true,
          "triggers": [
            "unusual amount: Z-score 8.01"
          ]
        },
        "frequency_check": {
          "confidence_score": 10,
          "is_suspicious": false,
          "triggers": []
        },
        "pattern_check": {
          "confidence_score": 72.75922721998579,
          "is_suspicious": false,
          "triggers": []
        }
      }
    },
    {
      "transaction_id": "0df020eb-8bb3-3536-869a-7bc850a9acbf",
      "fraud_score": 61.64954819617092,
      "risk_level": "medium",
      "detection_results": {
        "amount_check": {
          "confidence_score": 100,
          "is_suspicious": true,
          "triggers": [
            "unusual amount: Z-score 8.00"
          ]
        },
        "frequency_check": {
          "confidence_score": 20,
          "is_suspicious": false,
          "triggers": []
        },
        "pattern_check": {
          "confidence_score": 78.83182732056973,
          "is_suspicious": true,
          "triggers": [
            "spike in spending: 315% increase"
          ]
        }
      }
    }
  ],
  "processing_metadata": {
    "duration_ms": 21,
    "parallel_tasks": {
      "amount_analysis_duration_ms": 0,
      "frequency_analysis_duration_ms": 0,
      "pattern_analysis_duration_ms": 0
    },
    "total_transactions_analyzed": 1000
  }
}
```

#### Stacks
1. Go (Fiber)
2. SQLC (PGX)
