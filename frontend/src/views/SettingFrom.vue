<template>
    <a-form :model="formState" :label-col="labelCol" :wrapper-col="wrapperCol">
        <a-form-item label="自动接受对局">
            <a-switch v-model:checked="formState.autoConfirm" />
        </a-form-item>
        <a-form-item label="自动选用">
            <a-select v-model:value="formState.autoPick" show-search placeholder="选择英雄" style="width: 200px"
                :options="formState.options" :filter-option="filterOption" allowClear @focus="handleFocus"
                @blur="handleBlur" @change="handleChange"></a-select>
        </a-form-item>
        <a-form-item label="自动禁用">
            <a-select v-model:value="formState.autoBan" show-search placeholder="选择英雄" style="width: 200px" allowClear
                :options="formState.options" :filter-option="filterOption" @focus="handleFocus" @blur="handleBlur"
                @change="handleChange"></a-select>
        </a-form-item>
        <a-form-item :wrapper-col="{ span: 14, offset: 4 }">
            <a-button type="primary" @click="onSubmit">更新配置</a-button>
        </a-form-item>
    </a-form>
</template>
<script>
import { getConfig, updateConfig } from '@/api/bog'
import { defineComponent, reactive, toRaw, onMounted } from 'vue';
import { message } from 'ant-design-vue';


export default defineComponent({
    setup() {
        const formState = reactive({
            autoConfirm: false,
            autoPick: 0,
            autoBan: 0,
            options: [],
        });
        onMounted(async () => {
            getConfig().then(res => {
                const result = res.data.game;
                formState.autoConfirm = result.auto_confirm;
                formState.autoPick = result.auto_pick;
                formState.autoBan = result.auto_ban;
            })
            fetch('/champion.json')
                .then(response => response.json())
                .then(res => {
                    const result = res.data;
                    const data = [];
                    for (var key in result) {
                        var value = result[key];
                        data.push({
                            value: Number(value.key),
                            label: value.title + '-' + value.name,
                        });
                        formState.options = data;
                    }
                })
                .catch(error => console.error('Error:', error));
        })

        const onSubmit = () => {
            message
                .loading('同步配置中')
                .then(
                    () => updateConfig(formState.autoConfirm, formState.autoPick, formState.autoBan).then(() => {
                        message.success('更新成功')
                    }),
                );

        };

        const handleChange = value => {
            console.log(`selected ${value}`);
        };

        const handleBlur = () => {
            console.log('blur');
        };

        const handleFocus = () => {
            console.log('focus');
        };

        const filterOption = (input, option) => {
            return option.label.indexOf(input) >= 0;
        };
        return {
            labelCol: {
                style: {
                    width: '150px',
                },
            },
            wrapperCol: {
                span: 14,
            },
            filterOption,
            handleBlur,
            handleFocus,
            handleChange,
            formState,
            onSubmit,
        };
    },

});
</script>