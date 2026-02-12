import enumData from './enum'
export default {
    map: enumData.dictAttr || {},
    feMap: enumData.dict || {},
    // 对外提供转换为键值对功能
    getDict: function (key) {
        var arr = this.map[key]
        var dict = {}
        for (var i = 0; i < arr.length; i++) {
            dict[arr[i].key + ''] = arr[i].value
        }
        return dict
    },

    getFeDict: function (key) {
        return this.feMap[key]
    },

    // 按原始数组形式对外
    getDictArr: function (key) {
        return this.map[key]
    }
}
