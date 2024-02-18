**Payment System Prototype**

Go Playground: https://go.dev/play/p/jDCSp-ugRZq

**English:** 
Create one or more "classes" that will emulate the operation of a basic payment system.

This payment system does not use a database or file system; all information "lives" in memory only during the operation of your program (class).

In the payment system, each client has a unique account number in the IBAN format used in the Republic of Belarus (28 alphanumeric characters). For example: “BY04CBDC36029110100040000000”.

When your program starts, there are no accounts except for two special accounts: a government account for "emission" (money creation) and an account where money is sent for "destruction". You should come up with their numbers.

The class should be able to perform the following actions (functionality):
- Display the number of the special account for "emission";
- Display the number of the special account for "destruction";
- Carry out emission by adding a specified amount to the "emission" account;
- Send a certain amount of money from a specified account to the "destruction" account;
- Open a new account; you can generate a random account number or use your algorithm, or use an account number generated outside your class simply as a parameter;
- Transfer a specified amount of money between two specified accounts; provide two versions of this command:
  1) with multiple parameters
  2) with a single parameter in JSON format (structure it as you see fit);
- Display a list of all accounts, including special ones, indicating the balance of funds on them and their status ("active" or "blocked"). The information should be displayed in JSON format.

You are not required to implement all the listed functionality; you may only implement part of it, sufficient to demonstrate your skills. For example, a junior developer might only develop the class interface (list of functions and their parameters) and describe call scenarios and expected results, without actual implementation.

Write a number of calls to your class to demonstrate the functionality of the payment system. Output information to the standard output device (console, browser log, etc.); user input is not required, i.e., you can simply "hardcode" some scenarios, calling functions in a certain order with specific parameters for account numbers and amounts.

Demonstration of handling exceptional situations (sending money to a non-existent account, dealing with a "blocked" account, insufficient funds on the account for a transaction, etc.) is welcome.

Writing unit tests is encouraged.

Describing usage scenarios for your class functions (for example, in the form of comments or mini-documentation) is welcomed.

Ensure you demonstrate your work on one of the online compilers and send a link. Here are examples of online compilers; Google will provide more:
- https://www.onlinegdb.com/
- https://www.programiz.com/golang/online-compiler/
- https://go.dev/play/
- https://onecompiler.com/
- https://replit.com/~

In these systems, you can usually select the programming language for your interview in the top-right or side-left menu.

If you have trouble placing it on an online compiler, then create a comprehensive instruction on how to run your code.

Additionally, uploading the code to GitHub or another version control system is encouraged.

If you have significant questions about the meaning of this task, you may write them via email through the contact person who sent you the assignment.

**Russian:**
Создайте один или несколько “классов”, которые будет эмулировать работу простейшей платежной системы.

Платежная система не использует ни базу данных, ни файловую систему, а вся информация “живет” в оперативной памяти только во время работы вашей программы (класса).

В платежной системе каждый клиент имеет какой-то уникальный номер счета в формате IBAN, используемом в Республике Беларусь (28 букво-цифр). Например: вида “BY04CBDC36029110100040000000”. 

При запуске вашей программы в системе еще нет ни одного счета, кроме двух специальных счетов: счет государства, на который осуществляется “эмиссия” денег; и счет на который отправляются деньги для “уничтожения”. Сами придумайте их номера.

Класс должен уметь делать следующие действия (функционал):
выводить номер специального счета для “эмиссии”;
выводить номер специального счета для “уничтожения”;
осуществлять эмиссию, по добавлению на счет “эмиссии” указанной суммы;
осуществлять отправку определенной суммы денег с указанного счета на счет “уничтожения”;
открывать новый счет, вы можете генерировать случайный номер счета или по вашему алгоритму, или использовать сгенерированный вне вашего класса номер счета просто как параметр;
осуществлять перевод заданной суммы денег между двумя указанными счетами; обеспечить два варианта данной команды: 
1) с несколькими параметрами
2) с единственным параметром в формате json (структуру придумайте сами);
выводить список всех счетов, включая специальные, с указанием остатка денежных средств на них и их статуса (“активен” или “заблокирован”). Выводить необходимо в формате json.


Вы не обязаны реализовывать весь перечисленный функционал, можете делать только его часть, достаточную для демонстрации ваших навыков. Например, junior разработчик может разработать только интерфейс “класса” (перечень функций и их параметров) и описать сценарии вызовов и ожидаемые результаты, без реальной реализации.

Напишите какое-то количество вызовов вашего класса, чтобы продемонстрировать работоспособность платежной системы. Информацию выводите в стандартное устройство вывода (консоль, лог браузера, …), от пользователя можно ничего не вводить, т.е. можете просто “захардкодить” какие-то сценарии, вызывая в определенном порядке функции с конкретными параметрами номеров счетов и сумм.

Приветствуется демонстрация обработки нестандартных ситуаций (отправка денег на несуществующий счет, работа с “заблокированным” счетом, нехватка денег на счете для операции и т.п.).

Приветствуется написание Unit-тестов.

Приветствуется описание вариантов использования функций вашего класса (например, в виде комментариев или мини документации).

Обязательно обеспечьте демонстрацию вашей работы на одном из онлайн компиляторов и пришлите ссылку. Вот примеры, онлайн компиляторов, Google предоставит больше:
https://www.onlinegdb.com/
https://www.programiz.com/golang/online-compiler/
https://go.dev/play/
https://onecompiler.com/
https://replit.com/~

В указанных системах обычно сверху-справа, либо сбоку-слева можно выбрать тот язык программирования, на который вы проходите собеседование.
Если вы не справились с размещением на онлайн компиляторе, тогда создайте исчерпывающую инструкцию как запустить ваш код.

Дополнительно приветствуется заливка кода на github или другую систему контроля исходного кода.

Если у вас возникли существенные вопросы по смыслу этого задания, можете их письменно по электронной почте задать через контактное лицо, кто отправлял вам задание.


