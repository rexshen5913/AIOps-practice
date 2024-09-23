# AIOps Function Calling
## 以 Kubernetes 為服務 backend 
### function calling 實踐
目前實現了下面三種 function
* modify_config 
* restart_service
* apply_manifest

當執行 modify_config 這 function 時，也會接續執行 restart_service，
確保 k8s 的服務有正確的更新了配置

### 更新新的 function calling
1. 在 get_tools() 裡新增所需的 function
```
{
    "type": "function",
    "function": {
        "name": "Your_func_name",
        "description": "詳細的 function 功能",
        "parameters": {
            "type": "object",
            "properties": {
                "arg1": {"type": "string", "description": "arg1 的解釋"},
                "arg2": {"type": "string", "description": "arg2 的解釋"},
            },
            "required": ["arg1", "arg2"],
        },
    },
},
```
2. 在 Class KubernetesClient 更新想要執行的動作
3. 更新 handle_tool_call() 裡的 available_functions，將 Chatgpt 在 tool_call 裡選擇的 function，與後續你想要執行的動作 mapping
