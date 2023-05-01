# Guide to catcha solvers
Для обхода капчи в вайпалке имеется несколько опций на выбор. 

- Сервисы антикапчи (RuCaptcha)
- Распознавание нейронкой

## Распознавание
Вайпалка поддерживает обращение к локальному сервису для решения капчи. API такой же, как и [здесь](https://huggingface.co/spaces/neuroliptica/dvatch_captcha_sneedium_kuklofork).

### Схема
`POST http://127.0.0.1:7860/api/predict`

`Request`

```json 
{
    "data": ["base64 image"]
}
```

`Response`

```json
{
    "data": ["value"],
    "duration": 0.2781828
}
```
Поле `duration` в схеме ответа не обязательно. Вы так же можете [сами задать свой адрес](https://github.com/neuroliptica/dollwipe/blob/main/captcha/ocr.go#L29) для обращений к API солвера. 
